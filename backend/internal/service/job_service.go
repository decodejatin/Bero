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
	ErrUnauthorizedAction = errors.New("unauthorized action")
	ErrInvalidStatus      = errors.New("invalid job status transition")
	ErrAlreadyAssigned    = errors.New("job already assigned")
)

// JobService defines job business logic
type JobService interface {
	// Client operations
	CreateJob(ctx context.Context, clientID string, req *CreateJobRequest) (*domain.Job, error)
	GetClientJobs(ctx context.Context, clientID string, limit, offset int) ([]domain.Job, error)
	CancelJob(ctx context.Context, clientID, jobID string) error

	// Worker operations
	GetAvailableJobs(ctx context.Context, locality string, category *domain.ServiceCategory, limit, offset int) ([]domain.Job, error)
	AcceptJob(ctx context.Context, workerID, jobID string, estimatedArrival int) error
	StartJob(ctx context.Context, workerID, jobID string) error
	CompleteJob(ctx context.Context, workerID, jobID string, req *CompleteJobRequest) error
	ConfirmCompletion(ctx context.Context, clientID, jobID string) error
	GetWorkerJobs(ctx context.Context, workerID string, status *domain.JobStatus, limit, offset int) ([]domain.Job, error)

	// Common
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)
}

// CreateJobRequest for creating a new job
type CreateJobRequest struct {
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Category          domain.ServiceCategory `json:"category"`
	ClientName        string                 `json:"client_name"`
	Address           string                 `json:"address"`
	Locality          string                 `json:"locality"`
	City              string                 `json:"city"`
	Pincode           string                 `json:"pincode"`
	Latitude          *float64               `json:"latitude,omitempty"`
	Longitude         *float64               `json:"longitude,omitempty"`
	EstimatedDuration int                    `json:"estimated_duration_minutes"`
	PaymentAmount     float64                `json:"payment_amount_rupees"`
	ScheduledDate     time.Time              `json:"scheduled_date"`
	ScheduledTimeSlot string                 `json:"scheduled_time_slot"`
	IsUrgent          bool                   `json:"is_urgent"`
	RequiredSkills    []string               `json:"required_skills"`
}

// CompleteJobRequest for completing a job
type CompleteJobRequest struct {
	WorkerNotes    *string  `json:"worker_notes,omitempty"`
	PhotoProofURLs []string `json:"photo_proof_urls"`
}

type jobService struct {
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

// NewJobService creates a new job service
func NewJobService(jobRepo repository.JobRepository, userRepo repository.UserRepository) JobService {
	return &jobService{jobRepo: jobRepo, userRepo: userRepo}
}

// populateWorkerDetails fills in worker information for jobs that have assigned workers
func (s *jobService) populateWorkerDetails(ctx context.Context, jobs []domain.Job) []domain.Job {
	for i := range jobs {
		if jobs[i].AssignedWorkerID != nil {
			// Get worker user info
			worker, err := s.userRepo.GetByID(ctx, *jobs[i].AssignedWorkerID)
			if err == nil && worker != nil {
				jobs[i].AssignedWorkerName = worker.FullName
				jobs[i].AssignedWorkerPhone = &worker.PhoneNumber
			}

			// Get worker profile for rating
			workerProfile, err := s.userRepo.GetWorkerProfile(ctx, *jobs[i].AssignedWorkerID)
			if err == nil && workerProfile != nil {
				jobs[i].AssignedWorkerRating = &workerProfile.RatingAvg
			}
		}
	}
	return jobs
}

func (s *jobService) CreateJob(ctx context.Context, clientID string, req *CreateJobRequest) (*domain.Job, error) {
	job := &domain.Job{
		ID:                    uuid.New().String(),
		Title:                 req.Title,
		Description:           req.Description,
		Category:              req.Category,
		Status:                domain.JobStatusOpen,
		ClientID:              clientID,
		ClientName:            req.ClientName,
		Address:               req.Address,
		Locality:              req.Locality,
		City:                  req.City,
		Pincode:               req.Pincode,
		Latitude:              req.Latitude,
		Longitude:             req.Longitude,
		EstimatedDurationMins: req.EstimatedDuration,
		PaymentAmountRupees:   req.PaymentAmount,
		ScheduledDate:         req.ScheduledDate,
		ScheduledTimeSlot:     req.ScheduledTimeSlot,
		IsUrgent:              req.IsUrgent,
		RequiredSkills:        req.RequiredSkills,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

func (s *jobService) GetClientJobs(ctx context.Context, clientID string, limit, offset int) ([]domain.Job, error) {
	jobs, err := s.jobRepo.GetByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, err
	}
	// Populate worker details for accepted jobs
	return s.populateWorkerDetails(ctx, jobs), nil
}

func (s *jobService) CancelJob(ctx context.Context, clientID, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return ErrJobNotFound
	}

	if job.ClientID != clientID {
		return ErrUnauthorizedAction
	}

	if job.Status != domain.JobStatusOpen && job.Status != domain.JobStatusAssigned {
		return ErrInvalidStatus
	}

	job.Status = domain.JobStatusCancelled
	return s.jobRepo.Update(ctx, job)
}

func (s *jobService) GetAvailableJobs(ctx context.Context, locality string, category *domain.ServiceCategory, limit, offset int) ([]domain.Job, error) {
	return s.jobRepo.GetOpenJobs(ctx, locality, category, limit, offset)
}

func (s *jobService) AcceptJob(ctx context.Context, workerID, jobID string, estimatedArrival int) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return ErrJobNotFound
	}

	if job.Status != domain.JobStatusOpen {
		return ErrAlreadyAssigned
	}

	// Create acceptance record
	acceptance := &domain.JobAcceptance{
		ID:                   uuid.New().String(),
		JobID:                jobID,
		WorkerID:             workerID,
		AcceptedAt:           time.Now(),
		EstimatedArrivalMins: estimatedArrival,
	}

	if err := s.jobRepo.CreateAcceptance(ctx, acceptance); err != nil {
		return err
	}

	return s.jobRepo.AssignWorker(ctx, jobID, workerID)
}

func (s *jobService) StartJob(ctx context.Context, workerID, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return ErrJobNotFound
	}

	if job.AssignedWorkerID == nil || *job.AssignedWorkerID != workerID {
		return ErrUnauthorizedAction
	}

	if job.Status != domain.JobStatusAssigned {
		return ErrInvalidStatus
	}

	job.Status = domain.JobStatusInProgress
	return s.jobRepo.Update(ctx, job)
}

func (s *jobService) CompleteJob(ctx context.Context, workerID, jobID string, req *CompleteJobRequest) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return ErrJobNotFound
	}

	if job.AssignedWorkerID == nil || *job.AssignedWorkerID != workerID {
		return ErrUnauthorizedAction
	}

	if job.Status != domain.JobStatusInProgress {
		return ErrInvalidStatus
	}

	// Create completion record
	completion := &domain.JobCompletion{
		ID:             uuid.New().String(),
		JobID:          jobID,
		WorkerID:       workerID,
		CompletedAt:    time.Now(),
		WorkerNotes:    req.WorkerNotes,
		PhotoProofURLs: req.PhotoProofURLs,
	}

	if err := s.jobRepo.CreateCompletion(ctx, completion); err != nil {
		return err
	}

	// Worker marks complete -> AWAITING_CONFIRMATION
	job.WorkerConfirmed = true
	job.Status = domain.JobStatusAwaitingConfirmation
	return s.jobRepo.Update(ctx, job)
}

func (s *jobService) ConfirmCompletion(ctx context.Context, clientID, jobID string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return ErrJobNotFound
	}

	// Only client who owns the job can confirm
	if job.ClientID != clientID {
		return ErrUnauthorizedAction
	}

	// Must be awaiting confirmation
	if job.Status != domain.JobStatusAwaitingConfirmation {
		return ErrInvalidStatus
	}

	// Client confirms
	job.ClientConfirmed = true
	job.Status = domain.JobStatusCompleted
	return s.jobRepo.Update(ctx, job)
}

func (s *jobService) GetWorkerJobs(ctx context.Context, workerID string, status *domain.JobStatus, limit, offset int) ([]domain.Job, error) {
	return s.jobRepo.GetByWorkerID(ctx, workerID, status, limit, offset)
}

func (s *jobService) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return s.jobRepo.GetByID(ctx, jobID)
}
