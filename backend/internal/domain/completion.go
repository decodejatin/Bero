package domain

import "time"

// =============================================================================
// Dual-Sided Completion & Mandatory Rating Types
// =============================================================================

// MutualRating is a per-side rating with unique constraint (rater_id, job_id).
type MutualRating struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	JobID       string    `json:"job_id" gorm:"uniqueIndex:idx_rater_job;not null"`
	RaterID     string    `json:"rater_id" gorm:"uniqueIndex:idx_rater_job;not null;index"`
	RateeID     string    `json:"ratee_id" gorm:"index;not null"`
	RaterRole   string    `json:"rater_role" gorm:"not null"`   // "worker" or "client"
	RatingValue int       `json:"rating_value" gorm:"not null"` // 1–5
	ReviewText  string    `json:"review_text" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides GORM table name.
func (MutualRating) TableName() string {
	return "mutual_ratings"
}

// CompletionStatus reports the dual-sided completion & rating state.
type CompletionStatus struct {
	JobID             string     `json:"job_id"`
	Status            JobStatus  `json:"status"`
	WorkerCompleted   bool       `json:"worker_completed"`
	ClientConfirmed   bool       `json:"client_confirmed"`
	WorkerRated       bool       `json:"worker_rated"`
	ClientRated       bool       `json:"client_rated"`
	FullyCompleted    bool       `json:"fully_completed"`
	RatingsPending    bool       `json:"ratings_pending"`
	WorkerCompletedAt *time.Time `json:"worker_completed_at,omitempty"`
	ClientConfirmedAt *time.Time `json:"client_confirmed_at,omitempty"`
}
