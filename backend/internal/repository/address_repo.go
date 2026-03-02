package repository

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

// AddressRepository defines address data operations
type AddressRepository interface {
	Create(ctx context.Context, address *domain.SavedAddress) error
	GetByUserID(ctx context.Context, userID string) ([]domain.SavedAddress, error)
	GetByID(ctx context.Context, id string) (*domain.SavedAddress, error)
	Update(ctx context.Context, address *domain.SavedAddress) error
	Delete(ctx context.Context, id string) error
	SetDefault(ctx context.Context, userID, addressID string) error
}

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, address *domain.SavedAddress) error {
	if address.ID == "" {
		address.ID = uuid.New().String()
	}

	// If this is the first address or marked as default, clear other defaults
	if address.IsDefault {
		r.db.WithContext(ctx).Model(&domain.SavedAddress{}).
			Where("user_id = ?", address.UserID).
			Update("is_default", false)
	}

	return r.db.WithContext(ctx).Create(address).Error
}

func (r *addressRepository) GetByUserID(ctx context.Context, userID string) ([]domain.SavedAddress, error) {
	var addresses []domain.SavedAddress
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses)
	return addresses, result.Error
}

func (r *addressRepository) GetByID(ctx context.Context, id string) (*domain.SavedAddress, error) {
	var address domain.SavedAddress
	result := r.db.WithContext(ctx).First(&address, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAddressNotFound
		}
		return nil, result.Error
	}
	return &address, nil
}

func (r *addressRepository) Update(ctx context.Context, address *domain.SavedAddress) error {
	// If setting as default, clear other defaults first
	if address.IsDefault {
		r.db.WithContext(ctx).Model(&domain.SavedAddress{}).
			Where("user_id = ? AND id != ?", address.UserID, address.ID).
			Update("is_default", false)
	}
	return r.db.WithContext(ctx).Save(address).Error
}

func (r *addressRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.SavedAddress{}, "id = ?", id).Error
}

func (r *addressRepository) SetDefault(ctx context.Context, userID, addressID string) error {
	// Clear all defaults
	if err := r.db.WithContext(ctx).Model(&domain.SavedAddress{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		return err
	}
	// Set the selected one
	return r.db.WithContext(ctx).Model(&domain.SavedAddress{}).
		Where("id = ? AND user_id = ?", addressID, userID).
		Update("is_default", true).Error
}
