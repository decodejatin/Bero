package service

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
)

// RatingService defines rating business logic
type RatingService interface {
	SubmitRating(ctx context.Context, jobID, userID string, rating int, review string, tags []string) error
	GetJobRatings(ctx context.Context, jobID string) (*JobRatingsResponse, error)
}

type JobRatingsResponse struct {
	JobID        string   `json:"job_id"`
	ClientRating *int     `json:"client_rating,omitempty"`
	ClientReview *string  `json:"client_review,omitempty"`
	WorkerRating *int     `json:"worker_rating,omitempty"`
	WorkerReview *string  `json:"worker_review,omitempty"`
	RatingTags   []string `json:"rating_tags,omitempty"`
}

type ratingService struct {
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

func NewRatingService(jobRepo repository.JobRepository, userRepo repository.UserRepository) RatingService {
	return &ratingService{
		jobRepo:  jobRepo,
		userRepo: userRepo,
	}
}

func (s *ratingService) SubmitRating(ctx context.Context, jobID, userID string, rating int, review string, tags []string) error {
	// Validate rating
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	// Get the job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return errors.New("job not found")
	}

	// Job must be completed
	if job.Status != domain.JobStatusCompleted {
		return errors.New("can only rate completed jobs")
	}

	// Get or create completion record
	completion, err := s.jobRepo.GetCompletion(ctx, jobID)
	if err != nil {
		return err
	}
	if completion == nil {
		return errors.New("job completion record not found")
	}

	// Determine if user is client or worker
	reviewWithTags := review
	if len(tags) > 0 {
		reviewWithTags = review + " [Tags: " + joinTags(tags) + "]"
	}

	if job.ClientID == userID {
		// Client rates worker
		completion.ClientRating = &rating
		completion.ClientReview = &reviewWithTags

		// Update worker's average rating
		if job.AssignedWorkerID != nil {
			workerProfile, err := s.userRepo.GetWorkerProfile(ctx, *job.AssignedWorkerID)
			if err == nil && workerProfile != nil {
				newCount := workerProfile.RatingCount + 1
				newAvg := ((workerProfile.RatingAvg * float64(workerProfile.RatingCount)) + float64(rating)) / float64(newCount)
				workerProfile.RatingAvg = newAvg
				workerProfile.RatingCount = newCount
				s.userRepo.UpdateWorkerProfile(ctx, workerProfile)
			}
		}
	} else if job.AssignedWorkerID != nil && *job.AssignedWorkerID == userID {
		// Worker rates client
		completion.WorkerRating = &rating
		completion.WorkerReview = &reviewWithTags
	} else {
		return errors.New("you are not a participant of this job")
	}

	return s.jobRepo.UpdateCompletion(ctx, completion)
}

func (s *ratingService) GetJobRatings(ctx context.Context, jobID string) (*JobRatingsResponse, error) {
	completion, err := s.jobRepo.GetCompletion(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if completion == nil {
		return nil, errors.New("no ratings found")
	}

	return &JobRatingsResponse{
		JobID:        jobID,
		ClientRating: completion.ClientRating,
		ClientReview: completion.ClientReview,
		WorkerRating: completion.WorkerRating,
		WorkerReview: completion.WorkerReview,
	}, nil
}

func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ", "
		}
		result += t
	}
	return result
}
