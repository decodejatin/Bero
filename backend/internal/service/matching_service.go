package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
)

// MatchingService is the pre-processing brain for the Hungarian algorithm.
// It handles:
//  1. Candidate filtering — spatial + skill + availability
//  2. Feature normalization — all scores normalized 0–1 dynamically
//  3. Weight matrix construction — bounded [numJobs × K] output
//
// Design constraints:
//   - Candidate filtering < 30ms (PostGIS GiST + GIN indexes)
//   - Matrix build < 10ms (in-memory scoring, no DB calls)
//   - No N² city-wide scans — radius-bounded per job
//   - Thread-safe for 10,000+ concurrent users
type MatchingService interface {
	// GetCandidateWorkers retrieves and scores candidates for a single job.
	GetCandidateWorkers(ctx context.Context, jobID string, limit int) (*domain.CandidateResponse, error)

	// BuildWeightMatrix computes the cost matrix for a batch of jobs.
	// Each row is a job, columns are its K nearest candidates.
	BuildWeightMatrix(ctx context.Context, jobIDs []string, candidateLimit int) (*domain.WeightMatrix, error)

	// GetWeights returns the current dynamic matching weights.
	GetWeights(ctx context.Context) (*domain.MatchingWeights, error)

	// UpdateWeights persists new weight values at runtime.
	UpdateWeights(ctx context.Context, weights *domain.MatchingWeights) error
}

type matchingService struct {
	jobRepo       repository.JobRepository
	candidateRepo repository.CandidateRepository
	configRepo    repository.MatchingConfigRepository
}

// NewMatchingService creates a new matching service.
func NewMatchingService(
	jobRepo repository.JobRepository,
	candidateRepo repository.CandidateRepository,
	configRepo repository.MatchingConfigRepository,
) MatchingService {
	return &matchingService{
		jobRepo:       jobRepo,
		candidateRepo: candidateRepo,
		configRepo:    configRepo,
	}
}

// =============================================================================
// Candidate Filtering
// =============================================================================

func (s *matchingService) GetCandidateWorkers(ctx context.Context, jobID string, limit int) (*domain.CandidateResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	// 1. Fetch job details for location + skill requirements
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	if job.Latitude == nil || job.Longitude == nil {
		return nil, fmt.Errorf("job %s has no location set", jobID)
	}

	// 2. Determine search radius (configurable per deployment)
	radiusMeters := 2000.0 // Default 2km

	// 3. Query candidates with spatial + skill filtering
	candidates, err := s.candidateRepo.FindCandidateWorkers(
		ctx,
		*job.Latitude, *job.Longitude,
		radiusMeters,
		string(job.Category),
		job.RequiredSkills,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("find candidates: %w", err)
	}

	// 4. Load dynamic weights and normalization stats (parallel-safe via cache)
	weights, err := s.configRepo.GetWeights(ctx)
	if err != nil {
		return nil, fmt.Errorf("load weights: %w", err)
	}

	repStats, err := s.configRepo.GetReputationStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("load reputation stats: %w", err)
	}

	waitStats, err := s.configRepo.GetAvgWaitTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("load wait stats: %w", err)
	}

	// 5. Score each candidate
	for i := range candidates {
		scoreCandidateInPlace(&candidates[i], job, weights, repStats, waitStats, radiusMeters)
	}

	return &domain.CandidateResponse{
		JobID:        jobID,
		Candidates:   candidates,
		Count:        len(candidates),
		RadiusMeters: radiusMeters,
		Weights:      *weights,
	}, nil
}

// =============================================================================
// Weight Matrix Builder
// =============================================================================

func (s *matchingService) BuildWeightMatrix(ctx context.Context, jobIDs []string, candidateLimit int) (*domain.WeightMatrix, error) {
	if candidateLimit <= 0 {
		candidateLimit = 20
	}

	numJobs := len(jobIDs)
	if numJobs == 0 {
		return &domain.WeightMatrix{
			Weights:    [][]float64{},
			JobIDs:     []string{},
			WorkerIDs:  [][]string{},
			NumJobs:    0,
			MaxWorkers: 0,
		}, nil
	}

	// Pre-fetch shared resources once (amortized across all jobs)
	weights, err := s.configRepo.GetWeights(ctx)
	if err != nil {
		return nil, fmt.Errorf("load weights: %w", err)
	}

	repStats, err := s.configRepo.GetReputationStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("load reputation stats: %w", err)
	}

	waitStats, err := s.configRepo.GetAvgWaitTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("load wait stats: %w", err)
	}

	// Build matrix row by row (each job gets its own candidate set)
	matrix := &domain.WeightMatrix{
		Weights:    make([][]float64, numJobs),
		JobIDs:     make([]string, numJobs),
		WorkerIDs:  make([][]string, numJobs),
		NumJobs:    numJobs,
		MaxWorkers: 0,
	}

	radiusMeters := 2000.0

	for i, jobID := range jobIDs {
		job, err := s.jobRepo.GetByID(ctx, jobID)
		if err != nil {
			// Skip jobs that can't be found
			matrix.JobIDs[i] = jobID
			matrix.Weights[i] = []float64{}
			matrix.WorkerIDs[i] = []string{}
			continue
		}

		if job.Latitude == nil || job.Longitude == nil {
			matrix.JobIDs[i] = jobID
			matrix.Weights[i] = []float64{}
			matrix.WorkerIDs[i] = []string{}
			continue
		}

		candidates, err := s.candidateRepo.FindCandidateWorkers(
			ctx,
			*job.Latitude, *job.Longitude,
			radiusMeters,
			string(job.Category),
			job.RequiredSkills,
			candidateLimit,
		)
		if err != nil {
			matrix.JobIDs[i] = jobID
			matrix.Weights[i] = []float64{}
			matrix.WorkerIDs[i] = []string{}
			continue
		}

		// Score candidates and populate matrix row
		row := make([]float64, len(candidates))
		workerIDs := make([]string, len(candidates))

		for j := range candidates {
			scoreCandidateInPlace(&candidates[j], job, weights, repStats, waitStats, radiusMeters)
			row[j] = candidates[j].TotalScore
			workerIDs[j] = candidates[j].WorkerID
		}

		matrix.JobIDs[i] = jobID
		matrix.Weights[i] = row
		matrix.WorkerIDs[i] = workerIDs

		if len(candidates) > matrix.MaxWorkers {
			matrix.MaxWorkers = len(candidates)
		}
	}

	return matrix, nil
}

// =============================================================================
// Weight Config
// =============================================================================

func (s *matchingService) GetWeights(ctx context.Context) (*domain.MatchingWeights, error) {
	return s.configRepo.GetWeights(ctx)
}

func (s *matchingService) UpdateWeights(ctx context.Context, weights *domain.MatchingWeights) error {
	return s.configRepo.UpdateWeights(ctx, weights)
}

// =============================================================================
// Feature Normalization & Scoring
// All features normalized to 0–1 range using live dynamic statistics.
// No hardcoded minima/maxima — values are derived from current DB state.
// =============================================================================

// scoreCandidateInPlace computes all normalized scores and the final weighted
// total score, mutating the candidate struct in place for zero-allocation.
func scoreCandidateInPlace(
	c *domain.WorkerCandidate,
	job *domain.Job,
	weights *domain.MatchingWeights,
	repStats *domain.ReputationStats,
	waitStats *domain.WaitTimeStats,
	maxRadius float64,
) {
	// 1. Distance Score: 1 - (distance / maxRadius), clamped [0, 1]
	//    Closer workers get higher scores.
	c.DistanceScore = normalizeDistance(c.DistanceMeters, maxRadius)

	// 2. Reputation Score: (rating - min) / (max - min), from live distribution
	c.ReputationScore = normalizeReputation(c.RatingAvg, repStats)

	// 3. Skill Match Score: intersection / required, 0–1 ratio
	//    Not binary — partial overlap scores proportionally.
	c.SkillMatchScore = computeSkillMatch(c.Skills, job.RequiredSkills, string(job.Category))

	// 4. Wait Time Penalty: based on how recently the worker was last active.
	//    Uses the moving average from completed jobs as normalization base.
	c.WaitTimePenalty = computeWaitTimePenalty(c, waitStats)

	// 5. Weighted total: λ1*D + λ2*R + λ3*S - λ4*W
	c.TotalScore = weights.DistanceWeight*c.DistanceScore +
		weights.ReputationWeight*c.ReputationScore +
		weights.SkillWeight*c.SkillMatchScore -
		weights.WaitPenaltyWeight*c.WaitTimePenalty

	// Clamp total to [0, 1]
	c.TotalScore = math.Max(0, math.Min(1, c.TotalScore))
}

// normalizeDistance: 1 - (d / maxRadius), clamped to [0, 1]
func normalizeDistance(distanceMeters, maxRadius float64) float64 {
	if maxRadius <= 0 {
		return 1.0
	}
	score := 1.0 - (distanceMeters / maxRadius)
	return math.Max(0, math.Min(1, score))
}

// normalizeReputation: min-max normalization from live worker distribution
func normalizeReputation(ratingAvg float64, stats *domain.ReputationStats) float64 {
	rng := stats.MaxRating - stats.MinRating
	if rng <= 0 {
		return 0.5 // All workers same rating
	}
	score := (ratingAvg - stats.MinRating) / rng
	return math.Max(0, math.Min(1, score))
}

// computeSkillMatch: intersection ratio, not binary.
// If worker has 2 out of 3 required skills → 0.67.
// Falls back to category match if no explicit skills required.
func computeSkillMatch(workerSkills, requiredSkills []string, category string) float64 {
	if len(requiredSkills) == 0 {
		// No specific skills required — check category match
		if category == "" {
			return 1.0 // No requirements at all
		}
		for _, s := range workerSkills {
			if strings.EqualFold(s, category) {
				return 1.0
			}
		}
		return 0.3 // Worker doesn't match category but was returned by spatial query
	}

	// Count intersection
	workerSet := make(map[string]bool, len(workerSkills))
	for _, s := range workerSkills {
		workerSet[strings.ToUpper(s)] = true
	}

	matches := 0
	for _, req := range requiredSkills {
		if workerSet[strings.ToUpper(req)] {
			matches++
		}
	}

	return float64(matches) / float64(len(requiredSkills))
}

// computeWaitTimePenalty uses tier as a proxy for expected wait time.
// Gold workers respond faster → lower penalty. This can be extended
// with actual per-worker response time tracking for ML-based tuning.
func computeWaitTimePenalty(c *domain.WorkerCandidate, waitStats *domain.WaitTimeStats) float64 {
	// Base wait estimate from tier (proxy until we track per-worker metrics)
	var estimatedWaitMinutes float64
	switch c.Tier {
	case "GOLD":
		estimatedWaitMinutes = waitStats.AvgWaitMinutes * 0.5
	case "SILVER":
		estimatedWaitMinutes = waitStats.AvgWaitMinutes * 0.75
	default:
		estimatedWaitMinutes = waitStats.AvgWaitMinutes
	}

	// Normalize: min(wait / avgWait, 1.0)
	penalty := estimatedWaitMinutes / waitStats.AvgWaitMinutes
	return math.Min(penalty, 1.0)
}

// =============================================================================
// Utility: Compute elapsed minutes since a timestamp
// Available for future per-worker wait time tracking
// =============================================================================

func minutesSince(t time.Time) float64 {
	return time.Since(t).Minutes()
}
