package domain

import "time"

// UserType represents the type of user
type UserType string

const (
	UserTypeNone   UserType = "NONE"
	UserTypeWorker UserType = "WORKER"
	UserTypeClient UserType = "CLIENT"
)

// KycStatus represents KYC verification status
type KycStatus string

const (
	KycStatusNone     KycStatus = "NONE"
	KycStatusPending  KycStatus = "PENDING"
	KycStatusVerified KycStatus = "VERIFIED"
	KycStatusRejected KycStatus = "REJECTED"
)

// WorkerTier represents streak-based tier levels
type WorkerTier string

const (
	WorkerTierBronze WorkerTier = "BRONZE" // 15% commission
	WorkerTierSilver WorkerTier = "SILVER" // 12% commission
	WorkerTierGold   WorkerTier = "GOLD"   // 10% commission + insurance
)

// User represents a user in the system
type User struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	PhoneNumber      string    `json:"phone_number" gorm:"uniqueIndex;not null"`
	FullName         *string   `json:"full_name,omitempty"`
	Email            *string   `json:"email,omitempty"`
	AadhaarKycStatus KycStatus `json:"aadhaar_kyc_status" gorm:"default:NONE"`
	UserType         UserType  `json:"user_type" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// WorkerProfile represents worker-specific attributes
type WorkerProfile struct {
	UserID              string     `json:"user_id" gorm:"primaryKey"`
	Skills              []string   `json:"skills" gorm:"serializer:json"`
	H3IndexRes9         *string    `json:"h3_index_res9,omitempty"`
	IsOnline            bool       `json:"is_online" gorm:"default:false"`
	Latitude            *float64   `json:"latitude,omitempty"`
	Longitude           *float64   `json:"longitude,omitempty"`
	IsAvailable         bool       `json:"is_available" gorm:"default:false"`
	H3Index             *string    `json:"h3_index,omitempty"`
	WalletBalanceMicros int64      `json:"wallet_balance_micros" gorm:"default:0"`
	RatingAvg           float64    `json:"rating_avg" gorm:"default:0"`
	RatingCount         int        `json:"rating_count" gorm:"default:0"`
	StreakCount         int        `json:"streak_count" gorm:"default:0"`
	LastActiveDate      *time.Time `json:"last_active_date,omitempty"`
	Tier                WorkerTier `json:"tier" gorm:"default:BRONZE"`
	VideoBioURL         *string    `json:"video_bio_url,omitempty"`

	// Bayesian Reputation (§6.1)
	// TrustScore is the Wilson Score lower confidence bound of the Beta posterior.
	// It penalizes uncertainty — a new worker scores ~0.5, not 5.0★/5.0★.
	TrustScore     float64 `json:"trust_score" gorm:"default:0.5"`
	BayesSuccesses int     `json:"-" gorm:"default:0"` // S: ratings ≥4★
	BayesFailures  int     `json:"-" gorm:"default:0"` // F: ratings ≤3★

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relation
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// GetCommissionRate returns commission rate based on tier
func (w *WorkerProfile) GetCommissionRate() float64 {
	switch w.Tier {
	case WorkerTierSilver:
		return 0.12
	case WorkerTierGold:
		return 0.10
	default:
		return 0.15
	}
}

// CanAcceptJobs checks if worker can accept jobs
func (w *WorkerProfile) CanAcceptJobs() bool {
	return w.WalletBalanceMicros >= -50000000
}

// ClientProfile represents client-specific attributes
type ClientProfile struct {
	UserID         string    `json:"user_id" gorm:"primaryKey"`
	CompanyName    *string   `json:"company_name,omitempty"`
	GSTNumber      *string   `json:"gst_number,omitempty"`
	DefaultAddress *string   `json:"default_address,omitempty"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relation
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
