package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

// LocationRepository defines geospatial data operations
type LocationRepository interface {
	// UpdateWorkerLocation updates a worker's lat/lon, geometry, and H3 index
	UpdateWorkerLocation(ctx context.Context, workerID string, lat, lon float64, h3Index string) error

	// SetWorkerAvailability toggles the is_available flag
	SetWorkerAvailability(ctx context.Context, workerID string, available bool) error

	// FindNearbyWorkers returns available workers within radiusMeters, ordered by distance
	FindNearbyWorkers(ctx context.Context, lat, lon, radiusMeters float64, limit int) ([]domain.NearbyWorker, error)

	// FindWorkersByH3Indexes returns available workers matching any of the given H3 indexes
	FindWorkersByH3Indexes(ctx context.Context, h3Indexes []string) ([]domain.NearbyWorker, error)

	// UpdateJobLocation sets the geometry point and H3 index for a job
	UpdateJobLocation(ctx context.Context, jobID string, lat, lon float64, h3Index string) error
}

type locationRepository struct {
	db *gorm.DB
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) UpdateWorkerLocation(ctx context.Context, workerID string, lat, lon float64, h3Index string) error {
	// Raw SQL to update location fields + PostGIS geometry in a single statement
	query := `
		UPDATE worker_profiles
		SET latitude     = @lat,
		    longitude    = @lon,
		    location     = ST_SetSRID(ST_MakePoint(@lon, @lat), 4326),
		    h3_index     = @h3,
		    is_available = TRUE,
		    updated_at   = NOW()
		WHERE user_id = @workerID
	`
	result := r.db.WithContext(ctx).Exec(query,
		sql.Named("lat", lat),
		sql.Named("lon", lon),
		sql.Named("h3", h3Index),
		sql.Named("workerID", workerID),
	)
	if result.Error != nil {
		return fmt.Errorf("update worker location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("worker profile not found: %s", workerID)
	}
	return nil
}

func (r *locationRepository) SetWorkerAvailability(ctx context.Context, workerID string, available bool) error {
	result := r.db.WithContext(ctx).
		Model(&domain.WorkerProfile{}).
		Where("user_id = ?", workerID).
		Updates(map[string]interface{}{
			"is_available": available,
		})
	if result.Error != nil {
		return fmt.Errorf("set worker availability: %w", result.Error)
	}
	return nil
}

func (r *locationRepository) FindNearbyWorkers(ctx context.Context, lat, lon, radiusMeters float64, limit int) ([]domain.NearbyWorker, error) {
	if limit <= 0 || limit > 20 {
		limit = 20
	}

	// ST_DWithin on geography cast for accurate meter-based distance
	// ST_Distance returns meters when using geography
	query := `
		SELECT
			wp.user_id                                              AS worker_id,
			COALESCE(u.full_name, '')                               AS name,
			wp.latitude,
			wp.longitude,
			ST_Distance(
				wp.location::geography,
				ST_SetSRID(ST_MakePoint(@lon, @lat), 4326)::geography
			)                                                       AS distance_meters,
			COALESCE(wp.h3_index, '')                               AS h3_index,
			wp.rating_avg,
			wp.tier
		FROM worker_profiles wp
		JOIN users u ON u.id = wp.user_id
		WHERE wp.is_available = TRUE
		  AND wp.location IS NOT NULL
		  AND ST_DWithin(
				wp.location::geography,
				ST_SetSRID(ST_MakePoint(@lon, @lat), 4326)::geography,
				@radius
			)
		ORDER BY distance_meters ASC
		LIMIT @lim
	`

	var workers []domain.NearbyWorker
	result := r.db.WithContext(ctx).Raw(query,
		sql.Named("lat", lat),
		sql.Named("lon", lon),
		sql.Named("radius", radiusMeters),
		sql.Named("lim", limit),
	).Scan(&workers)

	if result.Error != nil {
		return nil, fmt.Errorf("find nearby workers: %w", result.Error)
	}
	if workers == nil {
		workers = []domain.NearbyWorker{}
	}
	return workers, nil
}

func (r *locationRepository) FindWorkersByH3Indexes(ctx context.Context, h3Indexes []string) ([]domain.NearbyWorker, error) {
	if len(h3Indexes) == 0 {
		return []domain.NearbyWorker{}, nil
	}

	query := `
		SELECT
			wp.user_id                           AS worker_id,
			COALESCE(u.full_name, '')            AS name,
			wp.latitude,
			wp.longitude,
			0                                    AS distance_meters,
			COALESCE(wp.h3_index, '')            AS h3_index,
			wp.rating_avg,
			wp.tier
		FROM worker_profiles wp
		JOIN users u ON u.id = wp.user_id
		WHERE wp.is_available = TRUE
		  AND wp.h3_index IN ?
	`

	var workers []domain.NearbyWorker
	result := r.db.WithContext(ctx).Raw(query, h3Indexes).Scan(&workers)
	if result.Error != nil {
		return nil, fmt.Errorf("find workers by h3: %w", result.Error)
	}
	if workers == nil {
		workers = []domain.NearbyWorker{}
	}
	return workers, nil
}

func (r *locationRepository) UpdateJobLocation(ctx context.Context, jobID string, lat, lon float64, h3Index string) error {
	query := `
		UPDATE jobs
		SET location   = ST_SetSRID(ST_MakePoint(@lon, @lat), 4326),
		    h3_index   = @h3,
		    updated_at = NOW()
		WHERE id = @jobID
	`
	result := r.db.WithContext(ctx).Exec(query,
		sql.Named("lat", lat),
		sql.Named("lon", lon),
		sql.Named("h3", h3Index),
		sql.Named("jobID", jobID),
	)
	if result.Error != nil {
		return fmt.Errorf("update job location: %w", result.Error)
	}
	return nil
}
