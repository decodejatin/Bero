package matchmaker

import "time"

// MatchableWorker represents a worker available for matching.
type MatchableWorker struct {
	ID        string
	Latitude  float64
	Longitude float64
	Skills    []string
	RatingAvg float64
	IsOnline  bool
}

// MatchableJob represents a job awaiting assignment.
type MatchableJob struct {
	ID             string
	Latitude       float64
	Longitude      float64
	RequiredSkills []string
	Category       string
	CreatedAt      time.Time
}

// Assignment represents a single worker-to-job assignment.
type Assignment struct {
	WorkerID string
	JobID    string
	Weight   float64
}

// MatchResult holds the output of one matching round.
type MatchResult struct {
	Assignments []Assignment
	Timestamp   time.Time
	WindowID    int64 // monotonically increasing window counter
}

// MatchConfig holds the tunable parameters for the matching algorithm.
type MatchConfig struct {
	// Weight coefficients (α values)
	AlphaProximity  float64 `json:"alpha_proximity"`  // α1 — proximity weight
	AlphaReputation float64 `json:"alpha_reputation"` // α2 — reputation weight
	AlphaSkillMatch float64 `json:"alpha_skill_match"` // α3 — skill match weight
	AlphaWaitTime   float64 `json:"alpha_wait_time"`  // α4 — wait time penalty

	// Batching window
	WindowDurationSeconds int `json:"window_duration_seconds"` // sliding window size (default 30)

	// Proximity settings
	MaxDistanceKm float64 `json:"max_distance_km"` // jobs beyond this distance get 0 proximity score

	// Minimum weight threshold — assignments below this score are discarded
	MinWeightThreshold float64 `json:"min_weight_threshold"`
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() MatchConfig {
	return MatchConfig{
		AlphaProximity:        0.35,
		AlphaReputation:       0.20,
		AlphaSkillMatch:       0.30,
		AlphaWaitTime:         0.15,
		WindowDurationSeconds: 30,
		MaxDistanceKm:         25.0,
		MinWeightThreshold:    0.10,
	}
}
