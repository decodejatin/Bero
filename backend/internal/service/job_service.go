package service

import (
	"context"
	"errors"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrJobNotFound        = errors.New("job not found")
	ErrNotJobOwner        = errors.New("you are not the owner of this job")
	ErrJobNotOpen         = errors.New("job is not open")
	ErrJobNotAssigned     = errors.New("job is not assigned")
	ErrNotAssignedWorker  = errors.New("you are not the assigned worker")
	ErrInvalidTransition  = errors.New("invalid job status transition")
	ErrCannotAcceptOwnJob = errors.New("cannot accept your own job")
)

// JobService defines job business operations
type JobService interface {
	CreateJob(ctx context.Context, userID string, job *domain.Job) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	GetAvailableJobs(ctx context.Context, locality string, category *domain.ServiceCategory, limit, offset int) ([]domain.Job, error)
	GetMyJobs(ctx context.Context, userID, userType string, limit, offset int) ([]domain.Job, error)
	AcceptJob(ctx context.Context, jobID, workerID string) (*domain.Job, error)
	StartJob(ctx context.Context, jobID, workerID string) (*domain.Job, error)
	CompleteJob(ctx context.Context, jobID, workerID string, notes *string, photoURLs []string) (*domain.Job, error)
	ConfirmCompletion(ctx context.Context, jobID, userID string) (*domain.Job, error)
	CancelJob(ctx context.Context, jobID, userID string) (*domain.Job, error)
}

type jobService struct {
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

// NewJobService creates a new job service
func NewJobService(jobRepo repository.JobRepository, userRepo repository.UserRepository) JobService {
	return &jobService{jobRepo: jobRepo, userRepo: userRepo}
}

func (s *jobService) CreateJob(ctx context.Context, userID string, job *domain.Job) (*domain.Job, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	clientName := ""
	if user.FullName != nil {
		clientName = *user.FullName
	}

	job.ID = uuid.New().String()
	job.ClientID = userID
	job.ClientName = clientName
	job.Status = domain.JobStatusOpen
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrJobNotFound
	}

	// Populate assigned worker details if present
	if job.AssignedWorkerID != nil {
		worker, err := s.userRepo.GetByID(ctx, *job.AssignedWorkerID)
		if err == nil {
			job.AssignedWorkerName = worker.FullName
			job.AssignedWorkerPhone = &worker.PhoneNumber
		}
	}

	return job, nil
}

func (s *jobService) GetAvailableJobs(ctx context.Context, locality string, category *domain.ServiceCategory, limit, offset int) ([]domain.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.jobRepo.GetOpenJobs(ctx, locality, category, limit, offset)
}

func (s *jobService) GetMyJobs(ctx context.Context, userID, userType string, limit, offset int) ([]domain.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	if userType == "WORKER" {
		return s.jobRepo.GetByWorkerID(ctx, userID, nil, limit, offset)
	}
	return s.jobRepo.GetByClientID(ctx, userID, limit, offset)
}

func (s *jobService) AcceptJob(ctx context.Context, jobID, workerID string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}
	if job.Status != domain.JobStatusOpen {
		return nil, ErrJobNotOpen
	}
	if job.ClientID == workerID {
		return nil, ErrCannotAcceptOwnJob
	}

	// Assign the worker
	if err := s.jobRepo.AssignWorker(ctx, jobID, workerID); err != nil {
		return nil, err
	}

	// Create acceptance record
	acceptance := &domain.JobAcceptance{
		ID:                   uuid.New().String(),
		JobID:                jobID,
		WorkerID:             workerID,
		AcceptedAt:           time.Now(),
		EstimatedArrivalMins: 30,
	}
	if err := s.jobRepo.CreateAcceptance(ctx, acceptance); err != nil {
		return nil, err
	}

	// Mark worker as unavailable so other clients don't see them
	workerProfile, err := s.userRepo.GetWorkerProfile(ctx, workerID)
	if err == nil && workerProfile != nil {
		workerProfile.IsAvailable = false
		_ = s.userRepo.UpdateWorkerProfile(ctx, workerProfile)
	}

	return s.jobRepo.GetByID(ctx, jobID)
}

func (s *jobService) StartJob(ctx context.Context, jobID, workerID string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}
	if job.Status != domain.JobStatusAssigned {
		return nil, ErrJobNotAssigned
	}
	if job.AssignedWorkerID == nil || *job.AssignedWorkerID != workerID {
		return nil, ErrNotAssignedWorker
	}

	job.Status = domain.JobStatusInProgress
	job.UpdatedAt = time.Now()
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) CompleteJob(ctx context.Context, jobID, workerID string, notes *string, photoURLs []string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}
	if job.Status != domain.JobStatusInProgress {
		return nil, ErrInvalidTransition
	}
	if job.AssignedWorkerID == nil || *job.AssignedWorkerID != workerID {
		return nil, ErrNotAssignedWorker
	}

	// Create completion record
	completion := &domain.JobCompletion{
		ID:             uuid.New().String(),
		JobID:          jobID,
		WorkerID:       workerID,
		CompletedAt:    time.Now(),
		WorkerNotes:    notes,
		PhotoProofURLs: photoURLs,
	}
	if err := s.jobRepo.CreateCompletion(ctx, completion); err != nil {
		return nil, err
	}

	job.Status = domain.JobStatusAwaitingConfirmation
	job.WorkerConfirmed = true
	job.UpdatedAt = time.Now()
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) ConfirmCompletion(ctx context.Context, jobID, userID string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}
	if job.Status != domain.JobStatusAwaitingConfirmation {
		return nil, ErrInvalidTransition
	}
	if job.ClientID != userID {
		return nil, ErrNotJobOwner
	}

	job.ClientConfirmed = true
	job.Status = domain.JobStatusCompleted
	job.UpdatedAt = time.Now()
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobService) CancelJob(ctx context.Context, jobID, userID string) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}

	// Only the client or the assigned worker can cancel
	if job.ClientID != userID && (job.AssignedWorkerID == nil || *job.AssignedWorkerID != userID) {
		return nil, ErrNotJobOwner
	}

	// Can only cancel if not already completed
	if job.Status == domain.JobStatusCompleted || job.Status == domain.JobStatusCancelled {
		return nil, ErrInvalidTransition
	}

	job.Status = domain.JobStatusCancelled
	job.UpdatedAt = time.Now()
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, err
	}

	// Re-enable worker availability if a worker was assigned
	if job.AssignedWorkerID != nil {
		workerProfile, err := s.userRepo.GetWorkerProfile(ctx, *job.AssignedWorkerID)
		if err == nil && workerProfile != nil {
			workerProfile.IsAvailable = true
			_ = s.userRepo.UpdateWorkerProfile(ctx, workerProfile)
		}
	}

	return job, nil
}
