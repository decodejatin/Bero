package repository

import (
	"context"
	"errors"

	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

var (
	ErrJobNotFound = errors.New("job not found")
)

// JobRepository defines job data operations
type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	GetByID(ctx context.Context, id string) (*domain.Job, error)
	GetByClientID(ctx context.Context, clientID string, limit, offset int) ([]domain.Job, error)
	GetOpenJobs(ctx context.Context, locality string, category *domain.ServiceCategory, limit, offset int) ([]domain.Job, error)
	GetByWorkerID(ctx context.Context, workerID string, status *domain.JobStatus, limit, offset int) ([]domain.Job, error)
	Update(ctx context.Context, job *domain.Job) error
	Delete(ctx context.Context, id string) error
	
	// Job lifecycle
	AssignWorker(ctx context.Context, jobID, workerID string) error
	CreateAcceptance(ctx context.Context, acceptance *domain.JobAcceptance) error
	CreateCompletion(ctx context.Context, completion *domain.JobCompletion) error
	GetCompletion(ctx context.Context, jobID string) (*domain.JobCompletion, error)
}

type jobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *jobRepository) GetByID(ctx context.Context, id string) (*domain.Job, error) {
	var job domain.Job
	result := r.db.WithContext(ctx).First(&job, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, result.Error
	}
	return &job, nil
}

func (r *jobRepository) GetByClientID(ctx context.Context, clientID string, limit, offset int) ([]domain.Job, error) {
	var jobs []domain.Job
	result := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs)
	return jobs, result.Error
}

func (r *jobRepository) GetOpenJobs(ctx context.Context, locality string, category *domain.ServiceCategory, limit, offset int) ([]domain.Job, error) {
	var jobs []domain.Job
	query := r.db.WithContext(ctx).Where("status = ?", domain.JobStatusOpen)
	
	if locality != "" {
		query = query.Where("locality = ?", locality)
	}
	if category != nil {
		query = query.Where("category = ?", *category)
	}
	
	result := query.
		Order("is_urgent DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs)
	return jobs, result.Error
}

func (r *jobRepository) GetByWorkerID(ctx context.Context, workerID string, status *domain.JobStatus, limit, offset int) ([]domain.Job, error) {
	var jobs []domain.Job
	query := r.db.WithContext(ctx).Where("assigned_worker_id = ?", workerID)
	
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	
	result := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs)
	return jobs, result.Error
}

func (r *jobRepository) Update(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *jobRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Job{}, "id = ?", id).Error
}

func (r *jobRepository) AssignWorker(ctx context.Context, jobID, workerID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"assigned_worker_id": workerID,
			"status":             domain.JobStatusAssigned,
		}).Error
}

func (r *jobRepository) CreateAcceptance(ctx context.Context, acceptance *domain.JobAcceptance) error {
	return r.db.WithContext(ctx).Create(acceptance).Error
}

func (r *jobRepository) CreateCompletion(ctx context.Context, completion *domain.JobCompletion) error {
	return r.db.WithContext(ctx).Create(completion).Error
}

func (r *jobRepository) GetCompletion(ctx context.Context, jobID string) (*domain.JobCompletion, error) {
	var completion domain.JobCompletion
	result := r.db.WithContext(ctx).First(&completion, "job_id = ?", jobID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &completion, nil
}
