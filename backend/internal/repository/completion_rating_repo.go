package repository

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

// CompletionRatingRepository handles mutual_ratings table operations.
type CompletionRatingRepository struct {
	db *gorm.DB
}

// NewCompletionRatingRepository creates a new completion rating repository.
func NewCompletionRatingRepository(db *gorm.DB) *CompletionRatingRepository {
	return &CompletionRatingRepository{db: db}
}

// CreateRating inserts a new mutual rating.
func (r *CompletionRatingRepository) CreateRating(ctx context.Context, rating *domain.MutualRating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

// GetRating fetches a specific rating by rater and job.
func (r *CompletionRatingRepository) GetRating(ctx context.Context, raterID, jobID string) (*domain.MutualRating, error) {
	var rating domain.MutualRating
	result := r.db.WithContext(ctx).
		Where("rater_id = ? AND job_id = ?", raterID, jobID).
		First(&rating)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &rating, nil
}

// GetRatingsForJob fetches all ratings for a job.
func (r *CompletionRatingRepository) GetRatingsForJob(ctx context.Context, jobID string) ([]domain.MutualRating, error) {
	var ratings []domain.MutualRating
	result := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&ratings)
	return ratings, result.Error
}

// GetPendingRatingJobs returns job IDs where the user hasn't submitted their rating yet.
func (r *CompletionRatingRepository) GetPendingRatingJobs(ctx context.Context, userID, role string) ([]string, error) {
	var jobIDs []string

	// Find FULLY_COMPLETED jobs where this user is a participant but hasn't rated
	var query string
	var args []interface{}

	if role == "worker" {
		// Worker hasn't rated: job is FULLY_COMPLETED, worker assigned, worker_rated = false
		query = `
			SELECT j.id FROM jobs j
			WHERE j.assigned_worker_id = ?
			  AND j.status = ?
			  AND j.worker_rated = FALSE
			ORDER BY j.updated_at ASC
			LIMIT 5
		`
		args = []interface{}{userID, domain.JobStatusFullyCompleted}
	} else {
		// Client hasn't rated: job is FULLY_COMPLETED, client owns, client_rated = false
		query = `
			SELECT j.id FROM jobs j
			WHERE j.client_id = ?
			  AND j.status = ?
			  AND j.client_rated = FALSE
			ORDER BY j.updated_at ASC
			LIMIT 5
		`
		args = []interface{}{userID, domain.JobStatusFullyCompleted}
	}

	result := r.db.WithContext(ctx).Raw(query, args...).Scan(&jobIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobIDs, nil
}
