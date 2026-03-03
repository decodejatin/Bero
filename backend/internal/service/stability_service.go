package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/google/uuid"
)

// =============================================================================
// StabilityService — Enforces marketplace stability around the dispatch layer.
//
// Core responsibilities:
//   - Time-decaying utility computation
//   - Switching cost enforcement
//   - Blocking pair detection
//   - Cancellation rate limiting with cooldowns
//   - Stability analytics logging
//
// Thread-safe. Does NOT modify Hungarian logic.
// =============================================================================

// StabilityService defines the stability enforcement interface.
type StabilityService interface {
	// Utility computation
	ComputeUtility(jobValueRupees, distanceKm, ratingAvg, delayMinutes float64) *domain.UtilityScore

	// Switching cost enforcement
	CanReassign(ctx context.Context, workerID, currentJobID, newJobID string) (allowed bool, reason string, err error)

	// Blocking pair detection
	CheckBlockingPair(newUtility, currentUtility, switchCost float64) bool

	// Cancellation management
	RecordCancellation(ctx context.Context, userID, role, jobID, reason string) error
	CheckCancellationLimit(ctx context.Context, userID string) (*domain.CancellationStatus, error)

	// Config
	GetConfig(ctx context.Context) (*domain.StabilityConfig, error)
	UpdateConfig(ctx context.Context, config *domain.StabilityConfig) error

	// Analytics
	GetStabilityStats(ctx context.Context) (*domain.StabilityStats, error)
	GetUserStatus(ctx context.Context, userID string) (*domain.CancellationStatus, error)
}

type stabilityService struct {
	stabilityRepo repository.StabilityRepository
	jobRepo       repository.JobRepository
}

// NewStabilityService creates a new stability service.
func NewStabilityService(
	stabilityRepo repository.StabilityRepository,
	jobRepo repository.JobRepository,
) StabilityService {
	return &stabilityService{
		stabilityRepo: stabilityRepo,
		jobRepo:       jobRepo,
	}
}

// =============================================================================
// Utility Function
//
// U(w,j,t) = JobValue * exp(-λ * timeDelay) - TravelCost + ReputationScore
//
// Time-sensitive: stale assignments decay exponentially.
// =============================================================================

func (s *stabilityService) ComputeUtility(
	jobValueRupees, distanceKm, ratingAvg, delayMinutes float64,
) *domain.UtilityScore {
	// Load config for λ and travel cost rate
	cfg, _ := s.stabilityRepo.GetConfig(context.Background())
	if cfg == nil {
		cfg = &domain.StabilityConfig{
			DecayLambda:     0.01,
			TravelCostPerKm: 5.0,
		}
	}

	// Time decay: exp(-λ * minutes)
	timeDecay := math.Exp(-cfg.DecayLambda * delayMinutes)

	// Travel cost in rupees
	travelCost := distanceKm * cfg.TravelCostPerKm

	// Reputation bonus: normalized 0–1 (assuming 5-star scale)
	reputationBonus := ratingAvg / 5.0

	// Total utility
	totalUtility := jobValueRupees*timeDecay - travelCost + reputationBonus*100

	return &domain.UtilityScore{
		JobValue:        jobValueRupees,
		TimeDecay:       timeDecay,
		TravelCost:      travelCost,
		ReputationBonus: reputationBonus,
		TotalUtility:    totalUtility,
	}
}

// =============================================================================
// Blocking Pair Detection
//
// A matching is unstable if ∃ (worker w, job j) such that:
//   NewUtility(w,j) > CurrentUtility(w,current) + SwitchingCost
//   AND NewUtility(j,w) > CurrentUtility(j,current)
//
// Returns true if a blocking pair exists (i.e., switch is justified).
// =============================================================================

func (s *stabilityService) CheckBlockingPair(
	newUtility, currentUtility, switchCost float64,
) bool {
	return newUtility > (currentUtility + switchCost)
}

// =============================================================================
// Switching Cost Enforcement
//
// Reassignment is allowed ONLY if:
//   NewUtility > CurrentUtility + Cswitch
//
// Prevents chaotic reassignment loops.
// =============================================================================

func (s *stabilityService) CanReassign(
	ctx context.Context,
	workerID, currentJobID, newJobID string,
) (bool, string, error) {
	cfg, err := s.stabilityRepo.GetConfig(ctx)
	if err != nil {
		return false, "", fmt.Errorf("load stability config: %w", err)
	}

	// Check cancellation limit first
	status, err := s.CheckCancellationLimit(ctx, workerID)
	if err != nil {
		return false, "", err
	}
	if status.IsBlocked {
		reason := fmt.Sprintf("cancellation cooldown active until %v", status.CooldownEnds)
		// Log denied switch
		s.logEvent(ctx, domain.StabilityEventSwitchDenied, workerID, "worker", currentJobID,
			map[string]interface{}{"reason": "cooldown_active", "new_job": newJobID})
		return false, reason, nil
	}

	// Fetch both jobs to compute utilities
	currentJob, err := s.jobRepo.GetByID(ctx, currentJobID)
	if err != nil {
		return false, "current job not found", nil
	}
	newJob, err := s.jobRepo.GetByID(ctx, newJobID)
	if err != nil {
		return false, "new job not found", nil
	}

	// Compute utilities (simplified: use payment as job value, 0 distance for current)
	currentUtility := s.ComputeUtility(
		currentJob.PaymentAmountRupees, 0, 0,
		time.Since(currentJob.CreatedAt).Minutes(),
	)
	newUtility := s.ComputeUtility(
		newJob.PaymentAmountRupees, 0, 0,
		time.Since(newJob.CreatedAt).Minutes(),
	)

	// Log switch attempt
	s.logEvent(ctx, domain.StabilityEventSwitchAttempt, workerID, "worker", currentJobID,
		map[string]interface{}{
			"new_job":         newJobID,
			"current_utility": currentUtility.TotalUtility,
			"new_utility":     newUtility.TotalUtility,
			"switch_cost":     cfg.SwitchCostFixed,
		})

	// Check blocking pair condition
	if s.CheckBlockingPair(newUtility.TotalUtility, currentUtility.TotalUtility, cfg.SwitchCostFixed) {
		s.logEvent(ctx, domain.StabilityEventSwitchAllowed, workerID, "worker", currentJobID,
			map[string]interface{}{"new_job": newJobID})
		return true, "switch allowed: utility gain exceeds switching cost", nil
	}

	s.logEvent(ctx, domain.StabilityEventSwitchDenied, workerID, "worker", currentJobID,
		map[string]interface{}{
			"new_job":     newJobID,
			"utility_gap": newUtility.TotalUtility - currentUtility.TotalUtility,
			"switch_cost": cfg.SwitchCostFixed,
		})

	return false, "insufficient utility gain to justify switch", nil
}

// =============================================================================
// Cancellation Management
//
// - Max cancellations per hour (rate limiter)
// - Escalation flag after daily threshold
// - Automatic cooldown timer
// =============================================================================

func (s *stabilityService) RecordCancellation(
	ctx context.Context,
	userID, role, jobID, reason string,
) error {
	cfg, err := s.stabilityRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("load stability config: %w", err)
	}

	// Determine event type
	eventType := domain.StabilityEventWorkerCancel
	if role == "client" {
		eventType = domain.StabilityEventClientCancel
	}

	// Log the cancellation event
	if err := s.logEvent(ctx, eventType, userID, role, jobID,
		map[string]interface{}{"reason": reason}); err != nil {
		return err
	}

	// Check if this triggers a cooldown
	hourAgo := time.Now().Add(-1 * time.Hour)
	hourlyCount, err := s.stabilityRepo.GetCancellationCount(ctx, userID, hourAgo)
	if err != nil {
		return err
	}

	if hourlyCount >= cfg.MaxCancelsPerHour {
		// Trigger cooldown
		s.logEvent(ctx, domain.StabilityEventCooldownStart, userID, role, jobID,
			map[string]interface{}{
				"hourly_count": hourlyCount,
				"cooldown_min": cfg.CooldownMinutes,
			})
	}

	// Check daily escalation
	dailyCount, err := s.stabilityRepo.GetDailyCancellationCount(ctx, userID)
	if err != nil {
		return err
	}

	if dailyCount >= cfg.EscalationThreshold {
		s.logEvent(ctx, domain.StabilityEventEscalation, userID, role, jobID,
			map[string]interface{}{
				"daily_count": dailyCount,
				"threshold":   cfg.EscalationThreshold,
			})
	}

	return nil
}

func (s *stabilityService) CheckCancellationLimit(
	ctx context.Context,
	userID string,
) (*domain.CancellationStatus, error) {
	cfg, err := s.stabilityRepo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load stability config: %w", err)
	}

	hourAgo := time.Now().Add(-1 * time.Hour)
	hourlyCount, err := s.stabilityRepo.GetCancellationCount(ctx, userID, hourAgo)
	if err != nil {
		return nil, err
	}

	dailyCount, err := s.stabilityRepo.GetDailyCancellationCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	status := &domain.CancellationStatus{
		UserID:              userID,
		CancelsThisHour:     hourlyCount,
		CancelsToday:        dailyCount,
		MaxPerHour:          cfg.MaxCancelsPerHour,
		EscalationThreshold: cfg.EscalationThreshold,
		IsBlocked:           hourlyCount >= cfg.MaxCancelsPerHour,
		IsEscalated:         dailyCount >= cfg.EscalationThreshold,
	}

	// Compute cooldown end time if blocked
	if status.IsBlocked {
		cooldownEnd := time.Now().Add(time.Duration(cfg.CooldownMinutes) * time.Minute)
		status.CooldownEnds = &cooldownEnd
	}

	return status, nil
}

// =============================================================================
// Config & Analytics
// =============================================================================

func (s *stabilityService) GetConfig(ctx context.Context) (*domain.StabilityConfig, error) {
	return s.stabilityRepo.GetConfig(ctx)
}

func (s *stabilityService) UpdateConfig(ctx context.Context, config *domain.StabilityConfig) error {
	return s.stabilityRepo.UpdateConfig(ctx, config)
}

func (s *stabilityService) GetStabilityStats(ctx context.Context) (*domain.StabilityStats, error) {
	return s.stabilityRepo.GetStabilityStats(ctx)
}

func (s *stabilityService) GetUserStatus(ctx context.Context, userID string) (*domain.CancellationStatus, error) {
	return s.CheckCancellationLimit(ctx, userID)
}

// =============================================================================
// Helpers
// =============================================================================

func (s *stabilityService) logEvent(
	ctx context.Context,
	eventType domain.StabilityEventType,
	actorID, actorRole, jobID string,
	details map[string]interface{},
) error {
	detailsJSON, _ := json.Marshal(details)
	event := &domain.StabilityEvent{
		ID:        uuid.New().String(),
		EventType: eventType,
		ActorID:   actorID,
		ActorRole: actorRole,
		JobID:     jobID,
		Details:   string(detailsJSON),
	}
	return s.stabilityRepo.LogEvent(ctx, event)
}
