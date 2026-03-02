package service

import (
	"context"
	"errors"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotParticipant   = errors.New("user is not a participant in this conversation")
	ErrEmptyMessage     = errors.New("message content cannot be empty")
	ErrSelfConversation = errors.New("cannot create conversation with yourself")
)

// ChatService handles chat business logic
type ChatService interface {
	GetOrCreateConversation(ctx context.Context, currentUserID, participantID string, jobID *string) (*ConversationResponse, error)
	GetConversations(ctx context.Context, userID string) ([]ConversationResponse, error)
	SendMessage(ctx context.Context, conversationID, senderID, content, messageType string) (*domain.ChatMessage, error)
	GetMessages(ctx context.Context, conversationID, userID string, limit, offset int) ([]domain.ChatMessage, error)
	MarkAsRead(ctx context.Context, conversationID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

// ConversationResponse includes participant info for the frontend
type ConversationResponse struct {
	ID              string    `json:"id"`
	ParticipantID   string    `json:"participant_id"`
	ParticipantName string    `json:"participant_name"`
	JobID           *string   `json:"job_id,omitempty"`
	LastMessage     string    `json:"last_message"`
	LastMessageAt   time.Time `json:"last_message_at"`
	UnreadCount     int64     `json:"unread_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type chatService struct {
	chatRepo repository.ChatRepository
	userRepo repository.UserRepository
}

// NewChatService creates a new chat service
func NewChatService(chatRepo repository.ChatRepository, userRepo repository.UserRepository) ChatService {
	return &chatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

func (s *chatService) GetOrCreateConversation(ctx context.Context, currentUserID, participantID string, jobID *string) (*ConversationResponse, error) {
	if currentUserID == participantID {
		return nil, ErrSelfConversation
	}

	conv, err := s.chatRepo.GetOrCreateConversation(ctx, currentUserID, participantID, jobID)
	if err != nil {
		return nil, err
	}

	return s.toConversationResponse(ctx, conv, currentUserID)
}

func (s *chatService) GetConversations(ctx context.Context, userID string) ([]ConversationResponse, error) {
	convs, err := s.chatRepo.GetConversationsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responses []ConversationResponse
	for _, conv := range convs {
		resp, err := s.toConversationResponse(ctx, &conv, userID)
		if err != nil {
			continue // Skip conversations with missing user data
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

func (s *chatService) SendMessage(ctx context.Context, conversationID, senderID, content, messageType string) (*domain.ChatMessage, error) {
	if content == "" {
		return nil, ErrEmptyMessage
	}

	if messageType == "" {
		messageType = "text"
	}

	// Verify conversation exists and user is a participant
	conv, err := s.chatRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if conv.User1ID != senderID && conv.User2ID != senderID {
		return nil, ErrNotParticipant
	}

	msg := &domain.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    messageType,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := s.chatRepo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *chatService) GetMessages(ctx context.Context, conversationID, userID string, limit, offset int) ([]domain.ChatMessage, error) {
	// Verify user is a participant
	conv, err := s.chatRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if conv.User1ID != userID && conv.User2ID != userID {
		return nil, ErrNotParticipant
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.chatRepo.GetMessagesByConversation(ctx, conversationID, limit, offset)
}

func (s *chatService) MarkAsRead(ctx context.Context, conversationID, userID string) error {
	// Verify user is a participant
	conv, err := s.chatRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if conv.User1ID != userID && conv.User2ID != userID {
		return ErrNotParticipant
	}

	return s.chatRepo.MarkMessagesAsRead(ctx, conversationID, userID)
}

func (s *chatService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.chatRepo.GetUnreadCount(ctx, userID)
}

// toConversationResponse converts a domain conversation to API response with participant info
func (s *chatService) toConversationResponse(ctx context.Context, conv *domain.ChatConversation, currentUserID string) (*ConversationResponse, error) {
	// Determine the other participant
	participantID := conv.User2ID
	if conv.User1ID != currentUserID {
		participantID = conv.User1ID
	}

	// Get participant name
	participant, err := s.userRepo.GetByID(ctx, participantID)
	participantName := "User"
	if err == nil && participant.FullName != nil {
		participantName = *participant.FullName
	}

	// Get unread count for this conversation
	unreadCount := int64(0)
	// Simple approximation: count unread messages in this conversation for current user
	// (In production, you'd want per-conversation unread counts)

	return &ConversationResponse{
		ID:              conv.ID,
		ParticipantID:   participantID,
		ParticipantName: participantName,
		JobID:           conv.JobID,
		LastMessage:     conv.LastMessage,
		LastMessageAt:   conv.LastMessageAt,
		UnreadCount:     unreadCount,
		CreatedAt:       conv.CreatedAt,
	}, nil
}
