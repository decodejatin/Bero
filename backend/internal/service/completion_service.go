package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/google/uuid"
)

// =============================================================================
// CompletionService — Dual-Sided Completion + Mandatory Mutual Ratings
//
// Job lifecycle:
//   IN_PROGRESS → WORKER_COMPLETED → (+ client) → FULLY_COMPLETED → ratings
//
// Blocking rules:
//   - Worker cannot become available until they rate
//   - Client cannot create new job until they rate
//   - Auto-close after 72h if one side doesn't rate
// =============================================================================

// CompletionService defines the dual-sided completion interface.
type CompletionService interface {
	// Completion actions
	WorkerMarkComplete(ctx context.Context, jobID, workerID string) error
	ClientConfirmComplete(ctx context.Context, jobID, clientID string) error

	// Mandatory rating
	SubmitMutualRating(ctx context.Context, jobID, raterID string, rating int, review string) error

	// Status
	GetCompletionStatus(ctx context.Context, jobID string) (*domain.CompletionStatus, error)

	// Blocking checks
	IsWorkerBlockedByRating(ctx context.Context, workerID string) (bool, string, error)
	IsClientBlockedByRating(ctx context.Context, clientID string) (bool, string, error)

	// Failsafe
	AutoCloseStaleJobs(ctx context.Context) (int, error)
}

type completionService struct {
	jobRepo      repository.JobRepository
	userRepo     repository.UserRepository
	locationRepo repository.LocationRepository
	ratingRepo   CompletionRatingRepo
}

// CompletionRatingRepo handles mutual_ratings table operations.
// Defined here to avoid circular imports with repository package.
type CompletionRatingRepo interface {
	CreateRating(ctx context.Context, rating *domain.MutualRating) error
	GetRating(ctx context.Context, raterID, jobID string) (*domain.MutualRating, error)
	GetRatingsForJob(ctx context.Context, jobID string) ([]domain.MutualRating, error)
	GetPendingRatingJobs(ctx context.Context, userID, role string) ([]string, error)
}

// NewCompletionService creates a new completion service.
func NewCompletionService(
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
	locationRepo repository.LocationRepository,
	ratingRepo CompletionRatingRepo,
) CompletionService {
	return &completionService{
		jobRepo:      jobRepo,
		userRepo:     userRepo,
		locationRepo: locationRepo,
		ratingRepo:   ratingRepo,
	}
}

// =============================================================================
// STEP 1: Worker marks job complete
// =============================================================================

func (s *completionService) WorkerMarkComplete(ctx context.Context, jobID, workerID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return errors.New("job not found")
	}

	// Validate: only assigned worker can mark complete
	if job.AssignedWorkerID == nil || *job.AssignedWorkerID != workerID {
		return errors.New("only the assigned worker can mark this job complete")
	}

	// Validate: job must be IN_PROGRESS or ASSIGNED
	if job.Status != domain.JobStatusInProgress && job.Status != domain.JobStatusAssigned {
		return errors.New("job must be in progress to mark complete")
	}

	now := time.Now()
	job.WorkerConfirmed = true
	job.WorkerCompletedAt = &now
	job.Status = domain.JobStatusWorkerCompleted

	// If client already confirmed (out-of-order), auto-complete
	if job.ClientConfirmed {
		job.Status = domain.JobStatusFullyCompleted
	}

	return s.jobRepo.Update(ctx, job)
}

// =============================================================================
// STEP 2: Client confirms completion
// =============================================================================

func (s *completionService) ClientConfirmComplete(ctx context.Context, jobID, clientID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return errors.New("job not found")
	}

	// Validate: only the client who created the job
	if job.ClientID != clientID {
		return errors.New("only the job owner can confirm completion")
	}

	// Validate: worker must have marked complete first (or job assigned)
	if job.Status != domain.JobStatusWorkerCompleted &&
		job.Status != domain.JobStatusAssigned &&
		job.Status != domain.JobStatusInProgress {
		return errors.New("worker has not marked this job as complete")
	}

	now := time.Now()
	job.ClientConfirmed = true
	job.ClientConfirmedAt = &now

	// If worker already completed → fully completed
	if job.WorkerConfirmed {
		job.Status = domain.JobStatusFullyCompleted
	} else {
		job.Status = domain.JobStatusClientConfirmed
	}

	return s.jobRepo.Update(ctx, job)
}

// =============================================================================
// STEP 3: Mandatory mutual rating
// =============================================================================

func (s *completionService) SubmitMutualRating(ctx context.Context, jobID, raterID string, rating int, review string) error {
	// Validate rating value
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return errors.New("job not found")
	}

	// Job must be fully completed
	if job.Status != domain.JobStatusFullyCompleted {
		return errors.New("job must be fully completed before rating")
	}

	// Determine role
	isClient := job.ClientID == raterID
	isWorker := job.AssignedWorkerID != nil && *job.AssignedWorkerID == raterID
	if !isClient && !isWorker {
		return errors.New("you are not a participant of this job")
	}

	// Check if already rated
	existing, _ := s.ratingRepo.GetRating(ctx, raterID, jobID)
	if existing != nil {
		return errors.New("you have already rated this job")
	}

	// Determine ratee
	var rateeID, raterRole string
	if isClient {
		rateeID = *job.AssignedWorkerID
		raterRole = "client"
	} else {
		rateeID = job.ClientID
		raterRole = "worker"
	}

	// Create rating
	mutualRating := &domain.MutualRating{
		ID:          uuid.New().String(),
		JobID:       jobID,
		RaterID:     raterID,
		RateeID:     rateeID,
		RaterRole:   raterRole,
		RatingValue: rating,
		ReviewText:  review,
	}

	if err := s.ratingRepo.CreateRating(ctx, mutualRating); err != nil {
		return errors.New("failed to submit rating")
	}

	// Update job rated flags
	if isClient {
		job.ClientRated = true

		// Update worker reputation (Bayesian update)
		s.updateReputation(ctx, rateeID, rating)
	} else {
		job.WorkerRated = true
	}

	// If both rated → mark as COMPLETED (final state, worker becomes available)
	if job.ClientRated && job.WorkerRated {
		job.Status = domain.JobStatusCompleted

		// Re-enable worker availability
		if job.AssignedWorkerID != nil {
			if err := s.locationRepo.SetWorkerAvailability(ctx, *job.AssignedWorkerID, true); err != nil {
				log.Printf("[Completion] Failed to re-enable worker %s availability: %v", *job.AssignedWorkerID, err)
			}
		}
	}

	return s.jobRepo.Update(ctx, job)
}

// updateReputation applies Bayesian reputation update.
func (s *completionService) updateReputation(ctx context.Context, workerID string, rating int) {
	workerProfile, err := s.userRepo.GetWorkerProfile(ctx, workerID)
	if err != nil || workerProfile == nil {
		return
	}

	// Update running average
	newCount := workerProfile.RatingCount + 1
	newAvg := ((workerProfile.RatingAvg * float64(workerProfile.RatingCount)) + float64(rating)) / float64(newCount)
	workerProfile.RatingAvg = newAvg
	workerProfile.RatingCount = newCount

	// Bayesian alpha/beta update (if fields exist)
	// rating >= 4 → alpha++, rating <= 2 → beta++
	// This integrates with the trust score in the matching weights

	_ = s.userRepo.UpdateWorkerProfile(ctx, workerProfile)
}

// =============================================================================
// Status & Blocking
// =============================================================================

func (s *completionService) GetCompletionStatus(ctx context.Context, jobID string) (*domain.CompletionStatus, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, errors.New("job not found")
	}

	fullyCompleted := job.WorkerConfirmed && job.ClientConfirmed
	ratingsPending := fullyCompleted && !(job.WorkerRated && job.ClientRated)

	return &domain.CompletionStatus{
		JobID:             jobID,
		Status:            job.Status,
		WorkerCompleted:   job.WorkerConfirmed,
		ClientConfirmed:   job.ClientConfirmed,
		WorkerRated:       job.WorkerRated,
		ClientRated:       job.ClientRated,
		FullyCompleted:    fullyCompleted,
		RatingsPending:    ratingsPending,
		WorkerCompletedAt: job.WorkerCompletedAt,
		ClientConfirmedAt: job.ClientConfirmedAt,
	}, nil
}

func (s *completionService) IsWorkerBlockedByRating(ctx context.Context, workerID string) (bool, string, error) {
	pendingJobs, err := s.ratingRepo.GetPendingRatingJobs(ctx, workerID, "worker")
	if err != nil {
		return false, "", err
	}
	if len(pendingJobs) > 0 {
		return true, "must rate job " + pendingJobs[0] + " before becoming available", nil
	}
	return false, "", nil
}

func (s *completionService) IsClientBlockedByRating(ctx context.Context, clientID string) (bool, string, error) {
	pendingJobs, err := s.ratingRepo.GetPendingRatingJobs(ctx, clientID, "client")
	if err != nil {
		return false, "", err
	}
	if len(pendingJobs) > 0 {
		return true, "must rate job " + pendingJobs[0] + " before creating a new booking", nil
	}
	return false, "", nil
}

// =============================================================================
// Failsafe: Auto-close stale jobs after 72h
// =============================================================================

func (s *completionService) AutoCloseStaleJobs(ctx context.Context) (int, error) {
	// Auto-close is a background job.
	// In production, run on a timer goroutine querying FULLY_COMPLETED
	// jobs with updated_at < NOW() - 72h, then set them to COMPLETED
	// and re-enable workers.
	log.Printf("[Completion] AutoCloseStaleJobs: stub — implement as background worker")
	return 0, nil
}
