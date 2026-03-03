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

// MatchingConfigRepository provides access to dynamic matching weights.
// Weights are cached in memory with a configurable TTL to avoid
// hitting the database on every candidate scoring call.
type MatchingConfigRepository interface {
	// GetWeights returns the current matching weights.
	// Uses in-memory cache with 30s TTL, falls back to env vars if DB empty.
	GetWeights(ctx context.Context) (*domain.MatchingWeights, error)

	// UpdateWeights persists new weight values and invalidates cache.
	UpdateWeights(ctx context.Context, weights *domain.MatchingWeights) error

	// GetReputationStats returns live min/max/avg rating for normalization.
	GetReputationStats(ctx context.Context) (*domain.ReputationStats, error)

	// GetAvgWaitTime returns the moving average wait time (minutes)
	// from recently completed jobs for wait-time penalty normalization.
	GetAvgWaitTime(ctx context.Context) (*domain.WaitTimeStats, error)
}

type matchingConfigRepository struct {
	db *gorm.DB

	// In-memory cache for weights (avoid DB hit on every scoring call)
	cacheMu      sync.RWMutex
	cachedWeight *domain.MatchingWeights
	cacheTime    time.Time
	cacheTTL     time.Duration
}

// NewMatchingConfigRepository creates a new matching config repository.
func NewMatchingConfigRepository(db *gorm.DB) MatchingConfigRepository {
	return &matchingConfigRepository{
		db:       db,
		cacheTTL: 30 * time.Second,
	}
}

func (r *matchingConfigRepository) GetWeights(ctx context.Context) (*domain.MatchingWeights, error) {
	// Check cache first (hot path — no lock contention for reads)
	r.cacheMu.RLock()
	if r.cachedWeight != nil && time.Since(r.cacheTime) < r.cacheTTL {
		w := *r.cachedWeight // copy
		r.cacheMu.RUnlock()
		return &w, nil
	}
	r.cacheMu.RUnlock()

	// Cache miss — load from database
	var weights domain.MatchingWeights
	result := r.db.WithContext(ctx).First(&weights)
	if result.Error != nil {
		// Fallback to environment variables
		w := r.loadFromEnv()
		return w, nil
	}

	// Update cache
	r.cacheMu.Lock()
	r.cachedWeight = &weights
	r.cacheTime = time.Now()
	r.cacheMu.Unlock()

	return &weights, nil
}

func (r *matchingConfigRepository) UpdateWeights(ctx context.Context, weights *domain.MatchingWeights) error {
	// Upsert: update the single config row (id=1) or create if missing
	result := r.db.WithContext(ctx).
		Where("id = ?", 1).
		Assign(map[string]interface{}{
			"distance_weight":     weights.DistanceWeight,
			"reputation_weight":   weights.ReputationWeight,
			"skill_weight":        weights.SkillWeight,
			"wait_penalty_weight": weights.WaitPenaltyWeight,
			"updated_at":          time.Now(),
		}).
		FirstOrCreate(weights)

	if result.Error != nil {
		return fmt.Errorf("update matching weights: %w", result.Error)
	}

	// Invalidate cache immediately
	r.cacheMu.Lock()
	r.cachedWeight = nil
	r.cacheMu.Unlock()

	return nil
}

func (r *matchingConfigRepository) GetReputationStats(ctx context.Context) (*domain.ReputationStats, error) {
	var stats domain.ReputationStats
	result := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(MIN(rating_avg), 0) AS min_rating,
			COALESCE(MAX(rating_avg), 5) AS max_rating,
			COALESCE(AVG(rating_avg), 3) AS avg_rating
		FROM worker_profiles
		WHERE is_available = TRUE
		  AND rating_count > 0
	`).Scan(&stats)

	if result.Error != nil {
		return nil, fmt.Errorf("get reputation stats: %w", result.Error)
	}

	// Guard against division by zero (all workers same rating)
	if stats.MaxRating == stats.MinRating {
		stats.MaxRating = stats.MinRating + 1.0
	}

	return &stats, nil
}

func (r *matchingConfigRepository) GetAvgWaitTime(ctx context.Context) (*domain.WaitTimeStats, error) {
	var stats domain.WaitTimeStats

	// Compute average wait time from the last 100 completed jobs
	// Wait time = time between job creation and job acceptance
	result := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(
			AVG(EXTRACT(EPOCH FROM (ja.accepted_at - j.created_at)) / 60.0),
			30.0
		) AS avg_wait_minutes
		FROM job_acceptances ja
		JOIN jobs j ON j.id = ja.job_id
		WHERE j.status IN ('ASSIGNED', 'IN_PROGRESS', 'COMPLETED')
		ORDER BY ja.accepted_at DESC
		LIMIT 100
	`).Scan(&stats)

	if result.Error != nil {
		return nil, fmt.Errorf("get avg wait time: %w", result.Error)
	}

	// Clamp to minimum 1 minute to avoid division issues
	if stats.AvgWaitMinutes < 1.0 {
		stats.AvgWaitMinutes = 1.0
	}

	return &stats, nil
}

// loadFromEnv returns weights from environment variables as fallback.
func (r *matchingConfigRepository) loadFromEnv() *domain.MatchingWeights {
	return &domain.MatchingWeights{
		DistanceWeight:    envFloat("MATCHING_DISTANCE_WEIGHT", 0.4),
		ReputationWeight:  envFloat("MATCHING_REPUTATION_WEIGHT", 0.3),
		SkillWeight:       envFloat("MATCHING_SKILL_WEIGHT", 0.2),
		WaitPenaltyWeight: envFloat("MATCHING_WAIT_PENALTY_WEIGHT", 0.1),
		UpdatedAt:         time.Now(),
	}
}

func envFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
