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

// StabilityRepository handles stability config, event logging, and analytics.
type StabilityRepository interface {
	// Config
	GetConfig(ctx context.Context) (*domain.StabilityConfig, error)
	UpdateConfig(ctx context.Context, config *domain.StabilityConfig) error

	// Event logging
	LogEvent(ctx context.Context, event *domain.StabilityEvent) error

	// Cancellation tracking
	GetCancellationCount(ctx context.Context, userID string, since time.Time) (int, error)
	GetDailyCancellationCount(ctx context.Context, userID string) (int, error)

	// Analytics
	GetStabilityStats(ctx context.Context) (*domain.StabilityStats, error)
	GetAvgStabilityTime(ctx context.Context) (float64, error)
}

type stabilityRepository struct {
	db *gorm.DB

	// Cached config (avoid DB hit on every check)
	cacheMu   sync.RWMutex
	cachedCfg *domain.StabilityConfig
	cacheTime time.Time
	cacheTTL  time.Duration
}

// NewStabilityRepository creates a new stability repository.
func NewStabilityRepository(db *gorm.DB) StabilityRepository {
	return &stabilityRepository{
		db:       db,
		cacheTTL: 30 * time.Second,
	}
}

func (r *stabilityRepository) GetConfig(ctx context.Context) (*domain.StabilityConfig, error) {
	// Check cache
	r.cacheMu.RLock()
	if r.cachedCfg != nil && time.Since(r.cacheTime) < r.cacheTTL {
		cfg := *r.cachedCfg
		r.cacheMu.RUnlock()
		return &cfg, nil
	}
	r.cacheMu.RUnlock()

	var config domain.StabilityConfig
	result := r.db.WithContext(ctx).First(&config)
	if result.Error != nil {
		// Fallback to env-based defaults
		return r.loadFromEnv(), nil
	}

	// Update cache
	r.cacheMu.Lock()
	r.cachedCfg = &config
	r.cacheTime = time.Now()
	r.cacheMu.Unlock()

	return &config, nil
}

func (r *stabilityRepository) UpdateConfig(ctx context.Context, config *domain.StabilityConfig) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", 1).
		Assign(map[string]interface{}{
			"switch_cost_fixed":        config.SwitchCostFixed,
			"earnings_penalty_percent": config.EarningsPenaltyPercent,
			"max_cancels_per_hour":     config.MaxCancelsPerHour,
			"escalation_threshold":     config.EscalationThreshold,
			"cooldown_minutes":         config.CooldownMinutes,
			"decay_lambda":             config.DecayLambda,
			"travel_cost_per_km":       config.TravelCostPerKm,
			"visibility_penalty":       config.VisibilityPenalty,
			"updated_at":               time.Now(),
		}).
		FirstOrCreate(config)

	if result.Error != nil {
		return fmt.Errorf("update stability config: %w", result.Error)
	}

	// Invalidate cache
	r.cacheMu.Lock()
	r.cachedCfg = nil
	r.cacheMu.Unlock()

	return nil
}

func (r *stabilityRepository) LogEvent(ctx context.Context, event *domain.StabilityEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *stabilityRepository) GetCancellationCount(ctx context.Context, userID string, since time.Time) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&domain.StabilityEvent{}).
		Where("actor_id = ? AND event_type IN (?, ?) AND created_at >= ?",
			userID,
			domain.StabilityEventWorkerCancel,
			domain.StabilityEventClientCancel,
			since,
		).Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("get cancellation count: %w", result.Error)
	}
	return int(count), nil
}

func (r *stabilityRepository) GetDailyCancellationCount(ctx context.Context, userID string) (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return r.GetCancellationCount(ctx, userID, today)
}

func (r *stabilityRepository) GetStabilityStats(ctx context.Context) (*domain.StabilityStats, error) {
	stats := &domain.StabilityStats{}
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	// Total cancellations (last 24h)
	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type IN (?, ?) AND created_at >= ?",
			domain.StabilityEventWorkerCancel, domain.StabilityEventClientCancel, last24h).
		Count(&stats.TotalCancellations)

	// Worker vs client cancellations
	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventWorkerCancel, last24h).
		Count(&stats.WorkerCancellations)

	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventClientCancel, last24h).
		Count(&stats.ClientCancellations)

	// Blocking pairs detected
	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventBlockingPair, last24h).
		Count(&stats.BlockingPairsDetected)

	// Switch attempts / allowed / denied
	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventSwitchAttempt, last24h).
		Count(&stats.SwitchAttempts)

	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventSwitchAllowed, last24h).
		Count(&stats.SwitchesAllowed)

	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventSwitchDenied, last24h).
		Count(&stats.SwitchesDenied)

	// Avg stability time
	avgMin, _ := r.GetAvgStabilityTime(ctx)
	stats.AvgStabilityMinutes = avgMin

	// Active cooldowns
	r.db.WithContext(ctx).Model(&domain.StabilityEvent{}).
		Where("event_type = ? AND created_at >= ?", domain.StabilityEventCooldownStart,
			now.Add(-30*time.Minute)).
		Count(&stats.ActiveCooldowns)

	return stats, nil
}

func (r *stabilityRepository) GetAvgStabilityTime(ctx context.Context) (float64, error) {
	// Average time between assignment and cancellation for cancelled jobs
	var avgMinutes float64
	result := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(
			AVG(EXTRACT(EPOCH FROM (se.created_at - ja.accepted_at)) / 60.0),
			0
		)
		FROM stability_events se
		JOIN job_acceptances ja ON ja.job_id = se.job_id
		WHERE se.event_type IN ('WORKER_CANCEL', 'CLIENT_CANCEL')
		  AND se.created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&avgMinutes)

	if result.Error != nil {
		return 0, result.Error
	}
	return avgMinutes, nil
}

func (r *stabilityRepository) loadFromEnv() *domain.StabilityConfig {
	return &domain.StabilityConfig{
		SwitchCostFixed:        envFloatStab("STABILITY_SWITCH_COST", 0.15),
		EarningsPenaltyPercent: envFloatStab("STABILITY_EARNINGS_PENALTY", 5.0),
		MaxCancelsPerHour:      envIntStab("STABILITY_MAX_CANCELS_HOUR", 3),
		EscalationThreshold:    envIntStab("STABILITY_ESCALATION_THRESHOLD", 5),
		CooldownMinutes:        envIntStab("STABILITY_COOLDOWN_MINUTES", 30),
		DecayLambda:            envFloatStab("STABILITY_DECAY_LAMBDA", 0.01),
		TravelCostPerKm:        envFloatStab("STABILITY_TRAVEL_COST_KM", 5.0),
		VisibilityPenalty:      envFloatStab("STABILITY_VISIBILITY_PENALTY", 0.2),
	}
}

func envFloatStab(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func envIntStab(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
