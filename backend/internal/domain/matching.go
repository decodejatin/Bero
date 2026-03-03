package domain

import "time"

// =============================================================================
// Matching Domain Types
// Pre-processing layer for Batched Hungarian Algorithm
// =============================================================================

// WorkerCandidate is an enriched worker record with computed matching scores.
// Produced by the candidate filtering pipeline, consumed by the matrix builder.
type WorkerCandidate struct {
	WorkerID       string   `json:"worker_id"`
	Name           string   `json:"name"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	DistanceMeters float64  `json:"distance_meters"`
	H3Index        string   `json:"h3_index,omitempty"`
	RatingAvg      float64  `json:"rating_avg"`
	RatingCount    int      `json:"rating_count"`
	Skills         []string `json:"skills"`
	Tier           string   `json:"tier"`

	// Computed scores (0–1 normalized, dynamically derived)
	DistanceScore   float64 `json:"distance_score"`
	ReputationScore float64 `json:"reputation_score"`
	SkillMatchScore float64 `json:"skill_match_score"`
	WaitTimePenalty float64 `json:"wait_time_penalty"`
	TotalScore      float64 `json:"total_score"`
}

// CandidateFilter holds configurable parameters for candidate retrieval.
// No hardcoded defaults — values injected by service layer.
type CandidateFilter struct {
	RadiusMeters      float64  `json:"radius_meters"`
	Limit             int      `json:"limit"`
	RequireSkillMatch bool     `json:"require_skill_match"`
	Category          string   `json:"category,omitempty"`
	RequiredSkills    []string `json:"required_skills,omitempty"`
}

// MatchingWeights holds the dynamic λ weights for the scoring formula.
// Loaded from database at runtime — never hardcoded.
//
//	score = λ1*Distance + λ2*Reputation + λ3*SkillMatch - λ4*WaitPenalty
type MatchingWeights struct {
	ID                int       `json:"id" gorm:"primaryKey;autoIncrement"`
	DistanceWeight    float64   `json:"distance_weight" gorm:"not null;default:0.4"`
	ReputationWeight  float64   `json:"reputation_weight" gorm:"not null;default:0.3"`
	SkillWeight       float64   `json:"skill_weight" gorm:"not null;default:0.2"`
	WaitPenaltyWeight float64   `json:"wait_penalty_weight" gorm:"not null;default:0.1"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides GORM table name
func (MatchingWeights) TableName() string {
	return "matching_weights"
}

// ReputationStats holds min/max/avg for dynamic reputation normalization.
// Queried live from the worker_profiles table — no static constants.
type ReputationStats struct {
	MinRating float64 `json:"min_rating"`
	MaxRating float64 `json:"max_rating"`
	AvgRating float64 `json:"avg_rating"`
}

// WaitTimeStats holds average wait time for dynamic normalization.
type WaitTimeStats struct {
	AvgWaitMinutes float64 `json:"avg_wait_minutes"`
}

// WeightMatrix is the output consumed by the Hungarian algorithm.
// Dimensions: [numJobs][numCandidatesPerJob]
type WeightMatrix struct {
	Weights    [][]float64 `json:"weights"`
	JobIDs     []string    `json:"job_ids"`
	WorkerIDs  [][]string  `json:"worker_ids"` // Per-job candidate IDs (ragged)
	NumJobs    int         `json:"num_jobs"`
	MaxWorkers int         `json:"max_workers"`
}

// CandidateResponse wraps candidate results for API responses.
type CandidateResponse struct {
	JobID        string            `json:"job_id"`
	Candidates   []WorkerCandidate `json:"candidates"`
	Count        int               `json:"count"`
	RadiusMeters float64           `json:"radius_meters"`
	Weights      MatchingWeights   `json:"weights_used"`
}

// =============================================================================
// Hungarian Algorithm Output Types
// =============================================================================

// Assignment represents a single job→worker assignment from the Hungarian algorithm.
type Assignment struct {
	JobID      string    `json:"job_id"`
	WorkerID   string    `json:"worker_id"`
	Score      float64   `json:"score"`
	AssignedAt time.Time `json:"assigned_at"`
}

// BatchResult is the output of a single batch processing cycle.
type BatchResult struct {
	Assignments    []Assignment `json:"assignments"`
	UnmatchedJobs  []string     `json:"unmatched_jobs"`
	TotalJobs      int          `json:"total_jobs"`
	TotalMatched   int          `json:"total_matched"`
	ProcessingMs   int64        `json:"processing_ms"`
	BatchTimestamp time.Time    `json:"batch_timestamp"`
}

// BatchStatus reports the current state of the matching queue.
type BatchStatus struct {
	QueueDepth     int          `json:"queue_depth"`
	LastBatchAt    *time.Time   `json:"last_batch_at,omitempty"`
	LastResult     *BatchResult `json:"last_result,omitempty"`
	IsRunning      bool         `json:"is_running"`
	BatchIntervalS int          `json:"batch_interval_seconds"`
}
