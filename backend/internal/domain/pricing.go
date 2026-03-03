package domain

import "time"

// =============================================================================
// Dynamic Pricing Domain Types
// Per-H3-hexagon surge pricing with sigmoid market clearing.
// =============================================================================

// PricingConfig holds runtime-configurable pricing parameters.
// Stored in database, updatable without redeploy.
type PricingConfig struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// Sigmoid surge parameters: M(θ) = 1 + (Mmax-1) / (1 + e^(-k(θ-θ0)))
	MaxSurgeMultiplier    float64 `json:"max_surge_multiplier" gorm:"not null;default:3.0"`   // Mmax
	EquilibriumTheta      float64 `json:"equilibrium_theta" gorm:"not null;default:1.5"`      // θ0
	ElasticitySensitivity float64 `json:"elasticity_sensitivity" gorm:"not null;default:3.0"` // k

	// Emergency cap when supply = 0
	EmergencySurgeCap float64 `json:"emergency_surge_cap" gorm:"not null;default:2.5"`

	// Smoothing: max change in multiplier per 5 minutes
	MaxSurgeChangeRate float64 `json:"max_surge_change_rate" gorm:"not null;default:0.3"`

	// Minimum surge (never go below 1.0)
	MinSurgeMultiplier float64 `json:"min_surge_multiplier" gorm:"not null;default:1.0"`

	// kRing radius for supply counting (how many neighboring hexagons to include)
	SupplyKRingRadius int `json:"supply_kring_radius" gorm:"not null;default:1"`

	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides GORM table name.
func (PricingConfig) TableName() string {
	return "pricing_config"
}

// SurgeSnapshot represents real-time surge state for a single H3 hexagon.
type SurgeSnapshot struct {
	H3Index         string  `json:"h3_index"`
	Demand          int     `json:"demand"`           // Pending jobs in hex
	Supply          int     `json:"supply"`           // Available workers in hex + kRing
	Theta           float64 `json:"theta"`            // D/S
	SurgeMultiplier float64 `json:"surge_multiplier"` // M(θ)
	IsEmergency     bool    `json:"is_emergency"`     // S == 0
	SmoothedFrom    float64 `json:"smoothed_from"`    // Previous multiplier before smoothing
}

// SurgeHistory logs surge snapshots for analytics.
type SurgeHistory struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	H3Index         string    `json:"h3_index" gorm:"index;not null"`
	Demand          int       `json:"demand" gorm:"not null"`
	Supply          int       `json:"supply" gorm:"not null"`
	Theta           float64   `json:"theta" gorm:"not null"`
	SurgeMultiplier float64   `json:"surge_multiplier" gorm:"not null"`
	IsEmergency     bool      `json:"is_emergency" gorm:"default:false"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName overrides GORM table name.
func (SurgeHistory) TableName() string {
	return "surge_history"
}

// PriceQuote is the output of surge pricing computation.
type PriceQuote struct {
	JobID           string        `json:"job_id"`
	H3Index         string        `json:"h3_index"`
	BasePrice       float64       `json:"base_price"`
	SurgeMultiplier float64       `json:"surge_multiplier"`
	FinalPrice      float64       `json:"final_price"`
	Surge           SurgeSnapshot `json:"surge_details"`
}
