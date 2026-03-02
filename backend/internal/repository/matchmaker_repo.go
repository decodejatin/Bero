package repository

import (
	"context"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

// MatchmakerRepository defines data operations needed by the matching engine.
type MatchmakerRepository interface {
	// GetOnlineWorkers returns online workers, optionally filtered by skills.
	GetOnlineWorkers(ctx context.Context, skills []string) ([]domain.WorkerProfile, error)

	// GetUnmatchedOpenJobs returns OPEN jobs that have no assigned worker,
	// created after windowStart.
	GetUnmatchedOpenJobs(ctx context.Context, windowStart time.Time) ([]domain.Job, error)
}

type matchmakerRepository struct {
	db *gorm.DB
}

// NewMatchmakerRepository creates a new matchmaker repository.
func NewMatchmakerRepository(db *gorm.DB) MatchmakerRepository {
	return &matchmakerRepository{db: db}
}

func (r *matchmakerRepository) GetOnlineWorkers(ctx context.Context, skills []string) ([]domain.WorkerProfile, error) {
	var workers []domain.WorkerProfile

	query := r.db.WithContext(ctx).
		Preload("User").
		Where("is_online = ?", true)

	if len(skills) > 0 {
		// Filter workers that have at least one matching skill.
		// Since skills is stored as JSON, we use a raw condition.
		// This works with PostgreSQL JSON containment.
		for _, skill := range skills {
			query = query.Where("skills::text LIKE ?", "%"+skill+"%")
		}
	}

	result := query.Find(&workers)
	return workers, result.Error
}

func (r *matchmakerRepository) GetUnmatchedOpenJobs(ctx context.Context, windowStart time.Time) ([]domain.Job, error) {
	var jobs []domain.Job
	result := r.db.WithContext(ctx).
		Where("status = ? AND assigned_worker_id IS NULL AND created_at >= ?",
			domain.JobStatusOpen, windowStart).
		Order("is_urgent DESC, created_at ASC").
		Find(&jobs)
	return jobs, result.Error
}
