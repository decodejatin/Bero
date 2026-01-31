package repository

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

// UserRepository defines user data operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error

	// Profile methods
	GetWorkerProfile(ctx context.Context, userID string) (*domain.WorkerProfile, error)
	CreateWorkerProfile(ctx context.Context, profile *domain.WorkerProfile) error
	UpdateWorkerProfile(ctx context.Context, profile *domain.WorkerProfile) error
	GetClientProfile(ctx context.Context, userID string) (*domain.ClientProfile, error)
	CreateClientProfile(ctx context.Context, profile *domain.ClientProfile) error
	UpdateClientProfile(ctx context.Context, profile *domain.ClientProfile) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserExists
		}
		return result.Error
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, "phone_number = ?", phone)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Save(user)
	return result.Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id)
	return result.Error
}

// Profile methods

func (r *userRepository) GetWorkerProfile(ctx context.Context, userID string) (*domain.WorkerProfile, error) {
	var profile domain.WorkerProfile
	result := r.db.WithContext(ctx).First(&profile, "user_id = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &profile, nil
}

func (r *userRepository) CreateWorkerProfile(ctx context.Context, profile *domain.WorkerProfile) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(profile)
	return result.Error
}

func (r *userRepository) UpdateWorkerProfile(ctx context.Context, profile *domain.WorkerProfile) error {
	result := r.db.WithContext(ctx).Save(profile)
	return result.Error
}

func (r *userRepository) GetClientProfile(ctx context.Context, userID string) (*domain.ClientProfile, error) {
	var profile domain.ClientProfile
	result := r.db.WithContext(ctx).First(&profile, "user_id = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &profile, nil
}

func (r *userRepository) CreateClientProfile(ctx context.Context, profile *domain.ClientProfile) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(profile)
	return result.Error
}

func (r *userRepository) UpdateClientProfile(ctx context.Context, profile *domain.ClientProfile) error {
	result := r.db.WithContext(ctx).Save(profile)
	return result.Error
}
