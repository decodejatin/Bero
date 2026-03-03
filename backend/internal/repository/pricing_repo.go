package repository

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

// PricingRepository handles pricing config, demand/supply counting, and surge history.
type PricingRepository interface {
	// Config
	GetConfig(ctx context.Context) (*domain.PricingConfig, error)
	UpdateConfig(ctx context.Context, config *domain.PricingConfig) error

	// Demand / Supply per hexagon
	CountPendingJobs(ctx context.Context, h3Index string) (int, error)
	CountAvailableWorkers(ctx context.Context, h3Indexes []string) (int, error)

	// Surge history
	LogSurge(ctx context.Context, history *domain.SurgeHistory) error
	GetLastSurge(ctx context.Context, h3Index string) (*domain.SurgeHistory, error)
	GetSurgeHistory(ctx context.Context, h3Index string, limit int) ([]domain.SurgeHistory, error)

	// Job surge snapshot
	UpdateJobSurge(ctx context.Context, jobID string, multiplier, surgePrice float64) error
}

type pricingRepository struct {
	db *gorm.DB

	cacheMu   sync.RWMutex
	cachedCfg *domain.PricingConfig
	cacheTime time.Time
	cacheTTL  time.Duration
}

// NewPricingRepository creates a new pricing repository.
func NewPricingRepository(db *gorm.DB) PricingRepository {
	return &pricingRepository{
		db:       db,
		cacheTTL: 30 * time.Second,
	}
}

func (r *pricingRepository) GetConfig(ctx context.Context) (*domain.PricingConfig, error) {
	r.cacheMu.RLock()
	if r.cachedCfg != nil && time.Since(r.cacheTime) < r.cacheTTL {
		cfg := *r.cachedCfg
		r.cacheMu.RUnlock()
		return &cfg, nil
	}
	r.cacheMu.RUnlock()

	var config domain.PricingConfig
	result := r.db.WithContext(ctx).First(&config)
	if result.Error != nil {
		return r.loadFromEnv(), nil
	}

	r.cacheMu.Lock()
	r.cachedCfg = &config
	r.cacheTime = time.Now()
	r.cacheMu.Unlock()

	return &config, nil
}

func (r *pricingRepository) UpdateConfig(ctx context.Context, config *domain.PricingConfig) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", 1).
		Assign(map[string]interface{}{
			"max_surge_multiplier":   config.MaxSurgeMultiplier,
			"equilibrium_theta":      config.EquilibriumTheta,
			"elasticity_sensitivity": config.ElasticitySensitivity,
			"emergency_surge_cap":    config.EmergencySurgeCap,
			"max_surge_change_rate":  config.MaxSurgeChangeRate,
			"min_surge_multiplier":   config.MinSurgeMultiplier,
			"supply_kring_radius":    config.SupplyKRingRadius,
			"updated_at":             time.Now(),
		}).
		FirstOrCreate(config)

	if result.Error != nil {
		return fmt.Errorf("update pricing config: %w", result.Error)
	}

	r.cacheMu.Lock()
	r.cachedCfg = nil
	r.cacheMu.Unlock()

	return nil
}

// CountPendingJobs counts OPEN jobs in a specific H3 hexagon.
func (r *pricingRepository) CountPendingJobs(ctx context.Context, h3Index string) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Where("h3_index = ? AND status = ?", h3Index, domain.JobStatusOpen).
		Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("count pending jobs: %w", result.Error)
	}
	return int(count), nil
}

// CountAvailableWorkers counts available workers across given H3 hexagons (hex + kRing).
func (r *pricingRepository) CountAvailableWorkers(ctx context.Context, h3Indexes []string) (int, error) {
	if len(h3Indexes) == 0 {
		return 0, nil
	}
	var count int64
	result := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT wl.worker_id)
		FROM worker_locations wl
		JOIN worker_profiles wp ON wp.user_id = wl.worker_id
		WHERE wp.h3_index IN ?
		  AND wl.is_available = TRUE
		  AND wl.updated_at >= NOW() - INTERVAL '5 minutes'
	`, h3Indexes).Scan(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("count available workers: %w", result.Error)
	}
	return int(count), nil
}

func (r *pricingRepository) LogSurge(ctx context.Context, history *domain.SurgeHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *pricingRepository) GetLastSurge(ctx context.Context, h3Index string) (*domain.SurgeHistory, error) {
	var history domain.SurgeHistory
	result := r.db.WithContext(ctx).
		Where("h3_index = ?", h3Index).
		Order("created_at DESC").
		First(&history)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &history, nil
}

func (r *pricingRepository) GetSurgeHistory(ctx context.Context, h3Index string, limit int) ([]domain.SurgeHistory, error) {
	if limit <= 0 {
		limit = 50
	}
	var history []domain.SurgeHistory
	result := r.db.WithContext(ctx).
		Where("h3_index = ?", h3Index).
		Order("created_at DESC").
		Limit(limit).
		Find(&history)
	return history, result.Error
}

func (r *pricingRepository) UpdateJobSurge(ctx context.Context, jobID string, multiplier, surgePrice float64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"surge_multiplier": multiplier,
			"surge_price":      surgePrice,
		}).Error
}

func (r *pricingRepository) loadFromEnv() *domain.PricingConfig {
	return &domain.PricingConfig{
		MaxSurgeMultiplier:    envFloatPr("PRICING_MAX_SURGE", 3.0),
		EquilibriumTheta:      envFloatPr("PRICING_EQUILIBRIUM_THETA", 1.5),
		ElasticitySensitivity: envFloatPr("PRICING_ELASTICITY_K", 3.0),
		EmergencySurgeCap:     envFloatPr("PRICING_EMERGENCY_CAP", 2.5),
		MaxSurgeChangeRate:    envFloatPr("PRICING_MAX_CHANGE_RATE", 0.3),
		MinSurgeMultiplier:    envFloatPr("PRICING_MIN_SURGE", 1.0),
		SupplyKRingRadius:     envIntPr("PRICING_KRING_RADIUS", 1),
	}
}

func envFloatPr(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func envIntPr(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
