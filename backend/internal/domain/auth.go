package domain

import "time"

// OtpRequest represents an OTP request
type OtpRequest struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	PhoneNumber     string    `json:"phone_number" gorm:"index;not null"`
	OtpHash         string    `json:"-" gorm:"not null"` // Hashed OTP, never expose
	ExpiresAt       time.Time `json:"expires_at" gorm:"not null"`
	Verified        bool      `json:"verified" gorm:"default:false"`
	AttemptCount    int       `json:"-" gorm:"default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Session represents an authenticated session
type Session struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index;not null"`
	RefreshToken string    `json:"-" gorm:"uniqueIndex;not null"` // Hashed
	DeviceInfo   *string   `json:"device_info,omitempty"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// AuthTokens represents JWT tokens returned after authentication
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}
