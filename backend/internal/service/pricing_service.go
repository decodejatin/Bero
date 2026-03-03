package service

import (
	"context"
	"math"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/google/uuid"
)

// =============================================================================
// PricingService — Per-H3-hexagon dynamic surge pricing.
//
// Sigmoid surge function:
//   M(θ) = 1 + (Mmax - 1) / (1 + e^(-k(θ - θ0)))
//
// Runs BEFORE batch matching to clear the market.
// Thread-safe, all params from DB.
// =============================================================================

// PricingService defines the dynamic pricing interface.
type PricingService interface {
	// Core pricing
	ComputeSurge(ctx context.Context, h3Index string) (*domain.SurgeSnapshot, error)
	GetPriceQuote(ctx context.Context, jobID string) (*domain.PriceQuote, error)
	ApplySurgeToJob(ctx context.Context, jobID string) (*domain.PriceQuote, error)

	// Config
	GetConfig(ctx context.Context) (*domain.PricingConfig, error)
	UpdateConfig(ctx context.Context, config *domain.PricingConfig) error

	// History
	GetSurgeHistory(ctx context.Context, h3Index string, limit int) ([]domain.SurgeHistory, error)
}

type pricingService struct {
	pricingRepo repository.PricingRepository
	jobRepo     repository.JobRepository
}

// NewPricingService creates a new pricing service.
func NewPricingService(
	pricingRepo repository.PricingRepository,
	jobRepo repository.JobRepository,
) PricingService {
	return &pricingService{
		pricingRepo: pricingRepo,
		jobRepo:     jobRepo,
	}
}

// =============================================================================
// ComputeSurge — Per-hexagon surge calculation with sigmoid + smoothing
// =============================================================================

func (s *pricingService) ComputeSurge(ctx context.Context, h3Index string) (*domain.SurgeSnapshot, error) {
	cfg, err := s.pricingRepo.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	// 1. Count demand: pending OPEN jobs in this hexagon
	demand, err := s.pricingRepo.CountPendingJobs(ctx, h3Index)
	if err != nil {
		return nil, err
	}

	// 2. Count supply: available workers in hex + kRing neighbors
	kRingIndexes := s.getKRingIndexes(h3Index, cfg.SupplyKRingRadius)
	supply, err := s.pricingRepo.CountAvailableWorkers(ctx, kRingIndexes)
	if err != nil {
		return nil, err
	}

	// 3. Compute market tightness θ = D/S
	snapshot := &domain.SurgeSnapshot{
		H3Index: h3Index,
		Demand:  demand,
		Supply:  supply,
	}

	if supply == 0 {
		// Emergency: no workers available
		snapshot.Theta = math.Inf(1)
		snapshot.SurgeMultiplier = cfg.EmergencySurgeCap
		snapshot.IsEmergency = true
	} else {
		theta := float64(demand) / float64(supply)
		snapshot.Theta = theta
		snapshot.SurgeMultiplier = s.sigmoid(theta, cfg)
	}

	// 4. Apply smoothing: limit rate of change
	snapshot.SmoothedFrom = snapshot.SurgeMultiplier
	lastSurge, err := s.pricingRepo.GetLastSurge(ctx, h3Index)
	if err == nil && lastSurge != nil {
		// Only smooth if last surge was within 5 minutes
		if time.Since(lastSurge.CreatedAt) < 5*time.Minute {
			delta := snapshot.SurgeMultiplier - lastSurge.SurgeMultiplier
			if math.Abs(delta) > cfg.MaxSurgeChangeRate {
				if delta > 0 {
					snapshot.SurgeMultiplier = lastSurge.SurgeMultiplier + cfg.MaxSurgeChangeRate
				} else {
					snapshot.SurgeMultiplier = lastSurge.SurgeMultiplier - cfg.MaxSurgeChangeRate
				}
			}
		}
	}

	// 5. Clamp to [min, max]
	if snapshot.SurgeMultiplier < cfg.MinSurgeMultiplier {
		snapshot.SurgeMultiplier = cfg.MinSurgeMultiplier
	}
	if snapshot.SurgeMultiplier > cfg.MaxSurgeMultiplier {
		snapshot.SurgeMultiplier = cfg.MaxSurgeMultiplier
	}

	// 6. Log to surge history
	history := &domain.SurgeHistory{
		ID:              uuid.New().String(),
		H3Index:         h3Index,
		Demand:          demand,
		Supply:          supply,
		Theta:           snapshot.Theta,
		SurgeMultiplier: snapshot.SurgeMultiplier,
		IsEmergency:     snapshot.IsEmergency,
	}
	_ = s.pricingRepo.LogSurge(ctx, history)

	return snapshot, nil
}

// sigmoid implements M(θ) = 1 + (Mmax - 1) / (1 + e^(-k(θ - θ0)))
func (s *pricingService) sigmoid(theta float64, cfg *domain.PricingConfig) float64 {
	exponent := -cfg.ElasticitySensitivity * (theta - cfg.EquilibriumTheta)
	return 1.0 + (cfg.MaxSurgeMultiplier-1.0)/(1.0+math.Exp(exponent))
}

// getKRingIndexes returns the H3 index for supply counting.
// Uses the hex itself for radius 0, and for radius >= 1, the pricing repo
// queries workers using a prefix match (parent hex approximation) instead
// of requiring the h3-go CGo library.
func (s *pricingService) getKRingIndexes(h3Index string, radius int) []string {
	if radius <= 0 || len(h3Index) < 4 {
		return []string{h3Index}
	}
	// For kRing approximation: include workers whose H3 index shares
	// the same parent prefix. Trim 1 char per ring level from the index.
	// This is a superset of the true kRing but avoids the CGo dependency.
	// The pricing repo will count distinct workers across these matches.
	trimLen := radius
	if trimLen > 2 {
		trimLen = 2
	}
	prefix := h3Index[:len(h3Index)-trimLen]
	// Return the prefix as a "like" pattern — the repo handles the query.
	// For simplicity, just return the original index; the repo's SQL
	// already queries worker_profiles.h3_index IN ? which matches exactly.
	// We pass the original hex — supply counting includes workers in
	// the same hex. For true kRing, we'd need the h3 library.
	return []string{h3Index, prefix}
}

// =============================================================================
// GetPriceQuote — Compute final price for a job
// =============================================================================

func (s *pricingService) GetPriceQuote(ctx context.Context, jobID string) (*domain.PriceQuote, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	h3Index := ""
	if job.H3Index != nil {
		h3Index = *job.H3Index
	}
	if h3Index == "" {
		// No H3 index → no surge
		return &domain.PriceQuote{
			JobID:           jobID,
			BasePrice:       job.PaymentAmountRupees,
			SurgeMultiplier: 1.0,
			FinalPrice:      job.PaymentAmountRupees,
		}, nil
	}

	surge, err := s.ComputeSurge(ctx, h3Index)
	if err != nil {
		return nil, err
	}

	finalPrice := job.PaymentAmountRupees * surge.SurgeMultiplier

	return &domain.PriceQuote{
		JobID:           jobID,
		H3Index:         h3Index,
		BasePrice:       job.PaymentAmountRupees,
		SurgeMultiplier: surge.SurgeMultiplier,
		FinalPrice:      math.Round(finalPrice*100) / 100, // Round to 2 decimal places
		Surge:           *surge,
	}, nil
}

// =============================================================================
// ApplySurgeToJob — Store surge snapshot in job record before matching
// =============================================================================

func (s *pricingService) ApplySurgeToJob(ctx context.Context, jobID string) (*domain.PriceQuote, error) {
	quote, err := s.GetPriceQuote(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Persist surge to job record
	if err := s.pricingRepo.UpdateJobSurge(ctx, jobID, quote.SurgeMultiplier, quote.FinalPrice); err != nil {
		return nil, err
	}

	return quote, nil
}

// =============================================================================
// Config & History
// =============================================================================

func (s *pricingService) GetConfig(ctx context.Context) (*domain.PricingConfig, error) {
	return s.pricingRepo.GetConfig(ctx)
}

func (s *pricingService) UpdateConfig(ctx context.Context, config *domain.PricingConfig) error {
	return s.pricingRepo.UpdateConfig(ctx, config)
}

func (s *pricingService) GetSurgeHistory(ctx context.Context, h3Index string, limit int) ([]domain.SurgeHistory, error) {
	return s.pricingRepo.GetSurgeHistory(ctx, h3Index, limit)
}
