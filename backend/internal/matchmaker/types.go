package matchmaker

import "time"

// MatchableWorker represents a worker available for matching.
type MatchableWorker struct {
	ID           string
	Latitude     float64
	Longitude    float64
	H3Index      string // H3 cell index (resolution 9)
	Skills       []string
	RatingAvg    float64
	TrustScore   float64 // Bayesian Wilson Score (§6.1) — preferred over RatingAvg
	IsOnline     bool
	CurrentJobID string // currently assigned job (empty = unassigned)
}

// MatchableJob represents a job awaiting assignment.
type MatchableJob struct {
	ID             string
	Latitude       float64
	Longitude      float64
	H3Index        string // H3 cell index (resolution 9)
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
	AlphaProximity  float64 `json:"alpha_proximity"`   // α1 — proximity weight
	AlphaReputation float64 `json:"alpha_reputation"`  // α2 — reputation weight
	AlphaSkillMatch float64 `json:"alpha_skill_match"` // α3 — skill match weight
	AlphaWaitTime   float64 `json:"alpha_wait_time"`   // α4 — wait time penalty

	// Batching window
	WindowDurationSeconds int `json:"window_duration_seconds"` // sliding window size (default 30)

	// Proximity settings
	MaxDistanceKm float64 `json:"max_distance_km"` // jobs beyond this distance get 0 proximity score

	// Minimum weight threshold — assignments below this score are discarded
	MinWeightThreshold float64 `json:"min_weight_threshold"`

	// H3 Geospatial settings
	H3Resolution      int `json:"h3_resolution"`       // H3 resolution level (default 9)
	KNearestNeighbors int `json:"k_nearest_neighbors"` // candidates per job for pruning (default 10)

	// Enable dynamic candidate pruning (set false to use full matrix)
	EnablePruning bool `json:"enable_pruning"`

	// Stability parameters (§3.1.2 — Modified Gale-Shapley)
	LambdaDecay     float64 `json:"lambda_decay"`     // λ — time decay rate for utility function U(t)
	SwitchingCost   float64 `json:"switching_cost"`   // C_switch — cost to switch assignments
	EnableStability bool    `json:"enable_stability"` // enable stability enforcement post-matching

	// Operational patterns
	EnableShadowMode bool `json:"enable_shadow_mode"` // run shadow matching for A/B testing

	// Backpressure & rate limiting (the "Safety Valve")
	EnableRateLimiting     bool    `json:"enable_rate_limiting"`      // enable Token Bucket admission gate
	RateLimitPerSecond     float64 `json:"rate_limit_per_second"`     // token refill rate (default 2.0)
	RateLimitBurst         int     `json:"rate_limit_burst"`          // burst capacity (default 5)
	HighPressureQueueDepth int     `json:"high_pressure_queue_depth"` // trigger HIGH tier (default 10)
	CritPressureQueueDepth int     `json:"crit_pressure_queue_depth"` // trigger CRITICAL tier (default 50)
	HighPressureLatencyMs  int64   `json:"high_pressure_latency_ms"`  // EWMA latency for HIGH (default 500)
	CritPressureLatencyMs  int64   `json:"crit_pressure_latency_ms"`  // EWMA latency for CRITICAL (default 2000)
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
		H3Resolution:          9,
		KNearestNeighbors:     10,
		EnablePruning:         true,
		LambdaDecay:           0.05,
		SwitchingCost:         0.15,
		EnableStability:       true,
		EnableShadowMode:      false, // opt-in: set true to enable shadow A/B testing

		// Backpressure defaults — the Safety Valve
		EnableRateLimiting:     true,
		RateLimitPerSecond:     2.0, // 2 matching rounds/sec max sustained
		RateLimitBurst:         5,   // absorb short spikes of 5 triggers
		HighPressureQueueDepth: 10,
		CritPressureQueueDepth: 50,
		HighPressureLatencyMs:  500,
		CritPressureLatencyMs:  2000,
	}
}
