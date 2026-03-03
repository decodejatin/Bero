package domain

import "time"

// =============================================================================
// Stability Enforcement Domain Types
// Prevents reassignment abuse and enforces switching costs.
// =============================================================================

// StabilityEventType categorizes stability log entries.
type StabilityEventType string

const (
	StabilityEventWorkerCancel  StabilityEventType = "WORKER_CANCEL"
	StabilityEventClientCancel  StabilityEventType = "CLIENT_CANCEL"
	StabilityEventBlockingPair  StabilityEventType = "BLOCKING_PAIR"
	StabilityEventSwitchAttempt StabilityEventType = "SWITCH_ATTEMPT"
	StabilityEventSwitchAllowed StabilityEventType = "SWITCH_ALLOWED"
	StabilityEventSwitchDenied  StabilityEventType = "SWITCH_DENIED"
	StabilityEventCooldownStart StabilityEventType = "COOLDOWN_START"
	StabilityEventEscalation    StabilityEventType = "ESCALATION"
)

// StabilityConfig holds runtime-configurable stability parameters.
// Stored in database, updatable without redeploy.
type StabilityConfig struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// Switching cost: reassignment allowed only if NewUtility > CurrentUtility + SwitchCost
	SwitchCostFixed        float64 `json:"switch_cost_fixed" gorm:"not null;default:0.15"`
	EarningsPenaltyPercent float64 `json:"earnings_penalty_percent" gorm:"not null;default:5.0"`

	// Cancellation limits
	MaxCancelsPerHour   int `json:"max_cancels_per_hour" gorm:"not null;default:3"`
	EscalationThreshold int `json:"escalation_threshold" gorm:"not null;default:5"` // Per day
	CooldownMinutes     int `json:"cooldown_minutes" gorm:"not null;default:30"`

	// Utility function parameters
	DecayLambda     float64 `json:"decay_lambda" gorm:"not null;default:0.01"`      // exp(-λ * minutes)
	TravelCostPerKm float64 `json:"travel_cost_per_km" gorm:"not null;default:5.0"` // Rupees

	// Visibility penalty for repeat offenders
	VisibilityPenalty float64 `json:"visibility_penalty" gorm:"not null;default:0.2"` // 0–1 multiplier

	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides GORM table name.
func (StabilityConfig) TableName() string {
	return "stability_config"
}

// StabilityEvent is an analytics log entry for stability tracking.
type StabilityEvent struct {
	ID        string             `json:"id" gorm:"primaryKey"`
	EventType StabilityEventType `json:"event_type" gorm:"index;not null"`
	ActorID   string             `json:"actor_id" gorm:"index;not null"` // Worker or client ID
	ActorRole string             `json:"actor_role" gorm:"not null"`     // "worker" or "client"
	JobID     string             `json:"job_id" gorm:"index;not null"`
	Details   string             `json:"details" gorm:"type:text"` // JSON metadata
	CreatedAt time.Time          `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName overrides GORM table name.
func (StabilityEvent) TableName() string {
	return "stability_events"
}

// CancellationStatus reports a user's current cancellation standing.
type CancellationStatus struct {
	UserID              string     `json:"user_id"`
	CancelsThisHour     int        `json:"cancels_this_hour"`
	CancelsToday        int        `json:"cancels_today"`
	MaxPerHour          int        `json:"max_per_hour"`
	EscalationThreshold int        `json:"escalation_threshold"`
	IsBlocked           bool       `json:"is_blocked"`
	IsEscalated         bool       `json:"is_escalated"`
	CooldownEnds        *time.Time `json:"cooldown_ends,omitempty"`
}

// UtilityScore represents the computed utility for a worker-job pair.
type UtilityScore struct {
	WorkerID        string  `json:"worker_id"`
	JobID           string  `json:"job_id"`
	JobValue        float64 `json:"job_value"`
	TimeDecay       float64 `json:"time_decay"`
	TravelCost      float64 `json:"travel_cost"`
	ReputationBonus float64 `json:"reputation_bonus"`
	TotalUtility    float64 `json:"total_utility"`
}

// StabilityStats holds aggregate stability analytics.
type StabilityStats struct {
	TotalCancellations    int64   `json:"total_cancellations"`
	WorkerCancellations   int64   `json:"worker_cancellations"`
	ClientCancellations   int64   `json:"client_cancellations"`
	BlockingPairsDetected int64   `json:"blocking_pairs_detected"`
	SwitchAttempts        int64   `json:"switch_attempts"`
	SwitchesAllowed       int64   `json:"switches_allowed"`
	SwitchesDenied        int64   `json:"switches_denied"`
	AvgStabilityMinutes   float64 `json:"avg_stability_minutes"`
	ActiveCooldowns       int64   `json:"active_cooldowns"`
}
