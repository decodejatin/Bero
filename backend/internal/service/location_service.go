package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
)

var (
	ErrInvalidCoordinates = errors.New("latitude must be between -90 and 90, longitude between -180 and 180")
	ErrRadiusTooLarge     = errors.New("radius must be between 100 and 50000 meters")
	ErrWorkerNotFound     = errors.New("worker profile not found")
)

const (
	DefaultRadiusMeters = 2000.0
	MaxRadiusMeters     = 50000.0
	MinRadiusMeters     = 100.0
	MaxNearbyWorkers    = 20

	// H3 resolution 9: avg hex edge ~174m, area ~0.1 km²
	H3Resolution = 9
)

// LocationService defines location business operations.
// This layer will feed into the Hungarian matching system:
//   - UpdateWorkerLocation keeps worker positions fresh
//   - GetNearbyWorkers produces the candidate set for cost matrix rows
//   - H3 indexes enable fast cell-based pre-filtering
type LocationService interface {
	// UpdateWorkerLocation validates coordinates, computes H3 index, and persists
	UpdateWorkerLocation(ctx context.Context, workerID string, lat, lon float64) (h3Index string, err error)

	// GetNearbyWorkers returns available workers within radius, sorted by distance
	GetNearbyWorkers(ctx context.Context, lat, lon, radiusMeters float64) (*domain.NearbyWorkersResponse, error)

	// SetWorkerAvailability toggles the is_available flag
	SetWorkerAvailability(ctx context.Context, workerID string, available bool) error

	// SetWorkerOffline marks a worker as unavailable
	SetWorkerOffline(ctx context.Context, workerID string) error

	// UpdateJobLocation stores geometry and H3 index for a job after creation
	UpdateJobLocation(ctx context.Context, jobID string, lat, lon float64) error
}

type locationService struct {
	locationRepo repository.LocationRepository
}

// NewLocationService creates a new location service
func NewLocationService(locationRepo repository.LocationRepository) LocationService {
	return &locationService{locationRepo: locationRepo}
}

func (s *locationService) UpdateWorkerLocation(ctx context.Context, workerID string, lat, lon float64) (string, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return "", err
	}

	// Compute H3 index at resolution 9 (~174m edge)
	h3Index := latLonToH3(lat, lon, H3Resolution)

	if err := s.locationRepo.UpdateWorkerLocation(ctx, workerID, lat, lon, h3Index); err != nil {
		return "", fmt.Errorf("update worker location: %w", err)
	}

	return h3Index, nil
}

func (s *locationService) GetNearbyWorkers(ctx context.Context, lat, lon, radiusMeters float64) (*domain.NearbyWorkersResponse, error) {
	if err := validateCoordinates(lat, lon); err != nil {
		return nil, err
	}

	// Apply radius bounds
	if radiusMeters <= 0 {
		radiusMeters = DefaultRadiusMeters
	}
	if radiusMeters < MinRadiusMeters {
		radiusMeters = MinRadiusMeters
	}
	if radiusMeters > MaxRadiusMeters {
		radiusMeters = MaxRadiusMeters
	}

	workers, err := s.locationRepo.FindNearbyWorkers(ctx, lat, lon, radiusMeters, MaxNearbyWorkers)
	if err != nil {
		return nil, fmt.Errorf("get nearby workers: %w", err)
	}

	return &domain.NearbyWorkersResponse{
		Workers:      workers,
		QueryLat:     lat,
		QueryLon:     lon,
		RadiusMeters: radiusMeters,
		Count:        len(workers),
	}, nil
}

func (s *locationService) SetWorkerAvailability(ctx context.Context, workerID string, available bool) error {
	return s.locationRepo.SetWorkerAvailability(ctx, workerID, available)
}

func (s *locationService) SetWorkerOffline(ctx context.Context, workerID string) error {
	return s.locationRepo.SetWorkerAvailability(ctx, workerID, false)
}

func (s *locationService) UpdateJobLocation(ctx context.Context, jobID string, lat, lon float64) error {
	if err := validateCoordinates(lat, lon); err != nil {
		return err
	}
	h3Index := latLonToH3(lat, lon, H3Resolution)
	return s.locationRepo.UpdateJobLocation(ctx, jobID, lat, lon, h3Index)
}

// --- Helpers ---

func validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return ErrInvalidCoordinates
	}
	return nil
}

// latLonToH3 computes an H3 hex index without a CGO dependency.
// This is a lightweight approximation that produces deterministic,
// spatially-coherent cell IDs suitable for pre-filtering.
// For production kRing queries, integrate "github.com/uber/h3-go/v4".
func latLonToH3(lat, lon float64, resolution int) string {
	// Scale factor per resolution level (each level ~= 7x smaller area)
	scale := math.Pow(7, float64(resolution))

	// Normalize to positive coordinates
	normLat := (lat + 90.0) / 180.0
	normLon := (lon + 180.0) / 360.0

	// Grid cell indices
	cellLat := int(normLat * scale)
	cellLon := int(normLon * scale)

	return fmt.Sprintf("%x%x%x", resolution, cellLat, cellLon)
}
