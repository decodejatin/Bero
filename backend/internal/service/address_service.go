package service

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
)

// AddressService defines address business logic
type AddressService interface {
	GetAddresses(ctx context.Context, userID string) ([]domain.SavedAddress, error)
	CreateAddress(ctx context.Context, userID, label, fullAddress string, lat, lng *float64, isDefault bool) (*domain.SavedAddress, error)
	UpdateAddress(ctx context.Context, userID, addressID, label, fullAddress string, lat, lng *float64, isDefault bool) (*domain.SavedAddress, error)
	DeleteAddress(ctx context.Context, userID, addressID string) error
	SetDefault(ctx context.Context, userID, addressID string) error
}

type addressService struct {
	addressRepo repository.AddressRepository
}

func NewAddressService(addressRepo repository.AddressRepository) AddressService {
	return &addressService{addressRepo: addressRepo}
}

func (s *addressService) GetAddresses(ctx context.Context, userID string) ([]domain.SavedAddress, error) {
	return s.addressRepo.GetByUserID(ctx, userID)
}

func (s *addressService) CreateAddress(ctx context.Context, userID, label, fullAddress string, lat, lng *float64, isDefault bool) (*domain.SavedAddress, error) {
	address := &domain.SavedAddress{
		UserID:      userID,
		Label:       label,
		FullAddress: fullAddress,
		Latitude:    lat,
		Longitude:   lng,
		IsDefault:   isDefault,
	}

	if err := s.addressRepo.Create(ctx, address); err != nil {
		return nil, err
	}
	return address, nil
}

func (s *addressService) UpdateAddress(ctx context.Context, userID, addressID, label, fullAddress string, lat, lng *float64, isDefault bool) (*domain.SavedAddress, error) {
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	address.Label = label
	address.FullAddress = fullAddress
	address.Latitude = lat
	address.Longitude = lng
	address.IsDefault = isDefault

	if err := s.addressRepo.Update(ctx, address); err != nil {
		return nil, err
	}
	return address, nil
}

func (s *addressService) DeleteAddress(ctx context.Context, userID, addressID string) error {
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return err
	}

	if address.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.addressRepo.Delete(ctx, addressID)
}

func (s *addressService) SetDefault(ctx context.Context, userID, addressID string) error {
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return err
	}

	if address.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.addressRepo.SetDefault(ctx, userID, addressID)
}
