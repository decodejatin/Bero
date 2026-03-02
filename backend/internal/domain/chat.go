package domain

import "time"

// ChatConversation represents a chat thread between two users
type ChatConversation struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	User1ID       string    `json:"user1_id" gorm:"index;not null"`
	User2ID       string    `json:"user2_id" gorm:"index;not null"`
	JobID         *string   `json:"job_id,omitempty" gorm:"index"`
	LastMessage   string    `json:"last_message" gorm:"default:''"`
	LastMessageAt time.Time `json:"last_message_at" gorm:"autoCreateTime"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	User1 User `json:"-" gorm:"foreignKey:User1ID"`
	User2 User `json:"-" gorm:"foreignKey:User2ID"`
}

// ChatMessage represents a single message within a conversation
type ChatMessage struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ConversationID string    `json:"conversation_id" gorm:"index;not null"`
	SenderID       string    `json:"sender_id" gorm:"index;not null"`
	Content        string    `json:"content" gorm:"not null"`
	MessageType    string    `json:"message_type" gorm:"default:text"` // text, image, system
	IsRead         bool      `json:"is_read" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	Conversation ChatConversation `json:"-" gorm:"foreignKey:ConversationID"`
	Sender       User             `json:"-" gorm:"foreignKey:SenderID"`
}
