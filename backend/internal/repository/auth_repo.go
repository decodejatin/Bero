package repository

import (
	"context"
	"errors"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

var (
	ErrOtpNotFound = errors.New("otp request not found")
	ErrOtpExpired  = errors.New("otp expired")
	ErrTooManyAttempts = errors.New("too many attempts")
)

// AuthRepository defines auth data operations
type AuthRepository interface {
	CreateOtpRequest(ctx context.Context, req *domain.OtpRequest) error
	GetOtpRequest(ctx context.Context, id string) (*domain.OtpRequest, error)
	GetOtpByPhone(ctx context.Context, phone string) (*domain.OtpRequest, error)
	UpdateOtpRequest(ctx context.Context, req *domain.OtpRequest) error
	DeleteExpiredOtps(ctx context.Context) error
	
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSession(ctx context.Context, id string) (*domain.Session, error)
	GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*domain.Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteUserSessions(ctx context.Context, userID string) error
}

type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateOtpRequest(ctx context.Context, req *domain.OtpRequest) error {
	// Delete existing OTPs for this phone
	r.db.WithContext(ctx).Delete(&domain.OtpRequest{}, "phone_number = ?", req.PhoneNumber)
	
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *authRepository) GetOtpRequest(ctx context.Context, id string) (*domain.OtpRequest, error) {
	var req domain.OtpRequest
	result := r.db.WithContext(ctx).First(&req, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrOtpNotFound
		}
		return nil, result.Error
	}
	return &req, nil
}

func (r *authRepository) GetOtpByPhone(ctx context.Context, phone string) (*domain.OtpRequest, error) {
	var req domain.OtpRequest
	result := r.db.WithContext(ctx).
		Where("phone_number = ? AND expires_at > ? AND verified = ?", phone, time.Now(), false).
		Order("created_at DESC").
		First(&req)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrOtpNotFound
		}
		return nil, result.Error
	}
	return &req, nil
}

func (r *authRepository) UpdateOtpRequest(ctx context.Context, req *domain.OtpRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *authRepository) DeleteExpiredOtps(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&domain.OtpRequest{}, "expires_at < ?", time.Now()).Error
}

func (r *authRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *authRepository) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	result := r.db.WithContext(ctx).First(&session, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, result.Error
	}
	return &session, nil
}

func (r *authRepository) GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var session domain.Session
	result := r.db.WithContext(ctx).First(&session, "refresh_token = ?", tokenHash)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, result.Error
	}
	return &session, nil
}

func (r *authRepository) DeleteSession(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Session{}, "id = ?", id).Error
}

func (r *authRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&domain.Session{}, "user_id = ?", userID).Error
}
