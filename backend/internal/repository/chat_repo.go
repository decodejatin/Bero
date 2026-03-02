package repository

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
	ErrUnauthorized         = errors.New("unauthorized access")
)

// ChatRepository defines chat data operations
type ChatRepository interface {
	// Conversation methods
	GetOrCreateConversation(ctx context.Context, user1ID, user2ID string, jobID *string) (*domain.ChatConversation, error)
	GetConversationByID(ctx context.Context, id string) (*domain.ChatConversation, error)
	GetConversationsByUser(ctx context.Context, userID string) ([]domain.ChatConversation, error)

	// Message methods
	CreateMessage(ctx context.Context, msg *domain.ChatMessage) error
	GetMessagesByConversation(ctx context.Context, conversationID string, limit, offset int) ([]domain.ChatMessage, error)
	MarkMessagesAsRead(ctx context.Context, conversationID, readerID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

type chatRepository struct {
	db *gorm.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) GetOrCreateConversation(ctx context.Context, user1ID, user2ID string, jobID *string) (*domain.ChatConversation, error) {
	var conv domain.ChatConversation

	// Check if conversation already exists between these two users
	result := r.db.WithContext(ctx).Where(
		"(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		user1ID, user2ID, user2ID, user1ID,
	).First(&conv)

	if result.Error == nil {
		return &conv, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// Create new conversation
	conv = domain.ChatConversation{
		ID:      uuid.New().String(),
		User1ID: user1ID,
		User2ID: user2ID,
		JobID:   jobID,
	}

	if err := r.db.WithContext(ctx).Create(&conv).Error; err != nil {
		return nil, err
	}

	return &conv, nil
}

func (r *chatRepository) GetConversationByID(ctx context.Context, id string) (*domain.ChatConversation, error) {
	var conv domain.ChatConversation
	result := r.db.WithContext(ctx).First(&conv, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, result.Error
	}
	return &conv, nil
}

func (r *chatRepository) GetConversationsByUser(ctx context.Context, userID string) ([]domain.ChatConversation, error) {
	var convs []domain.ChatConversation
	result := r.db.WithContext(ctx).
		Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("last_message_at DESC").
		Find(&convs)
	if result.Error != nil {
		return nil, result.Error
	}
	return convs, nil
}

func (r *chatRepository) CreateMessage(ctx context.Context, msg *domain.ChatMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the message
		if err := tx.Create(msg).Error; err != nil {
			return err
		}

		// Update conversation's last message
		if err := tx.Model(&domain.ChatConversation{}).
			Where("id = ?", msg.ConversationID).
			Updates(map[string]interface{}{
				"last_message":    msg.Content,
				"last_message_at": msg.CreatedAt,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *chatRepository) GetMessagesByConversation(ctx context.Context, conversationID string, limit, offset int) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	result := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	return messages, nil
}

func (r *chatRepository) MarkMessagesAsRead(ctx context.Context, conversationID, readerID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.ChatMessage{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = false", conversationID, readerID).
		Update("is_read", true).Error
}

func (r *chatRepository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64

	// Get all conversations where user is a participant
	subQuery := r.db.WithContext(ctx).
		Model(&domain.ChatConversation{}).
		Select("id").
		Where("user1_id = ? OR user2_id = ?", userID, userID)

	err := r.db.WithContext(ctx).
		Model(&domain.ChatMessage{}).
		Where("conversation_id IN (?) AND sender_id != ? AND is_read = false", subQuery, userID).
		Count(&count).Error

	return count, err
}
