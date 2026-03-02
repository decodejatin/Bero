package domain

import "time"

// SavedAddress represents a user's saved address
type SavedAddress struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index;not null"`
	Label       string    `json:"label" gorm:"not null"` // Home, Work, Other
	FullAddress string    `json:"full_address" gorm:"not null"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
