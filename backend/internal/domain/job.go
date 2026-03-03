package domain

import "time"

// JobStatus represents job lifecycle status
type JobStatus string

const (
	JobStatusOpen                 JobStatus = "OPEN"
	JobStatusAssigned             JobStatus = "ASSIGNED"
	JobStatusInProgress           JobStatus = "IN_PROGRESS"
	JobStatusWorkerCompleted      JobStatus = "WORKER_COMPLETED"
	JobStatusClientConfirmed      JobStatus = "CLIENT_CONFIRMED"
	JobStatusFullyCompleted       JobStatus = "FULLY_COMPLETED"
	JobStatusAwaitingConfirmation JobStatus = "AWAITING_CONFIRMATION"
	JobStatusCompleted            JobStatus = "COMPLETED"
	JobStatusCancelled            JobStatus = "CANCELLED"
	JobStatusDisputed             JobStatus = "DISPUTED"
)

// ServiceCategory represents service categories
type ServiceCategory string

const (
	ServiceCategoryPlumbing        ServiceCategory = "PLUMBING"
	ServiceCategoryElectrical      ServiceCategory = "ELECTRICAL"
	ServiceCategoryCarpentry       ServiceCategory = "CARPENTRY"
	ServiceCategoryPainting        ServiceCategory = "PAINTING"
	ServiceCategoryCleaning        ServiceCategory = "CLEANING"
	ServiceCategoryACRepair        ServiceCategory = "AC_REPAIR"
	ServiceCategoryApplianceRepair ServiceCategory = "APPLIANCE_REPAIR"
	ServiceCategoryPestControl     ServiceCategory = "PEST_CONTROL"
	ServiceCategoryGardening       ServiceCategory = "GARDENING"
	ServiceCategoryOther           ServiceCategory = "OTHER"
)

// Job represents a job listing
type Job struct {
	ID                    string          `json:"id" gorm:"primaryKey"`
	Title                 string          `json:"title" gorm:"not null"`
	Description           string          `json:"description" gorm:"type:text"`
	Category              ServiceCategory `json:"category" gorm:"not null"`
	Status                JobStatus       `json:"status" gorm:"default:OPEN"`
	ClientID              string          `json:"client_id" gorm:"index;not null"`
	ClientName            string          `json:"client_name" gorm:"not null"`
	ClientPhone           *string         `json:"client_phone,omitempty"` // Only visible after acceptance
	Address               string          `json:"address" gorm:"not null"`
	Locality              string          `json:"locality" gorm:"index"`
	City                  string          `json:"city" gorm:"default:Delhi"`
	Pincode               string          `json:"pincode" gorm:"index"`
	Latitude              *float64        `json:"latitude,omitempty"`
	Longitude             *float64        `json:"longitude,omitempty"`
	H3Index               *string         `json:"h3_index,omitempty"`
	EstimatedDurationMins int             `json:"estimated_duration_minutes" gorm:"default:60"`
	PaymentAmountRupees   float64         `json:"payment_amount_rupees" gorm:"not null"`
	ScheduledDate         time.Time       `json:"scheduled_date" gorm:"not null"`
	ScheduledTimeSlot     string          `json:"scheduled_time_slot" gorm:"not null"`
	IsUrgent              bool            `json:"is_urgent" gorm:"default:false"`
	RequiredSkills        []string        `json:"required_skills" gorm:"serializer:json"`
	AssignedWorkerID      *string         `json:"assigned_worker_id,omitempty" gorm:"index"`
	AssignedWorkerName    *string         `json:"assigned_worker_name,omitempty" gorm:"-"`   // Not stored, populated at runtime
	AssignedWorkerPhone   *string         `json:"assigned_worker_phone,omitempty" gorm:"-"`  // Not stored, populated at runtime
	AssignedWorkerRating  *float64        `json:"assigned_worker_rating,omitempty" gorm:"-"` // Not stored, populated at runtime
	WorkerConfirmed       bool            `json:"worker_confirmed" gorm:"default:false"`
	ClientConfirmed       bool            `json:"client_confirmed" gorm:"default:false"`
	WorkerCompletedAt     *time.Time      `json:"worker_completed_at,omitempty"`
	ClientConfirmedAt     *time.Time      `json:"client_confirmed_at,omitempty"`
	WorkerRated           bool            `json:"worker_rated" gorm:"default:false"`
	ClientRated           bool            `json:"client_rated" gorm:"default:false"`
	SurgeMultiplier       *float64        `json:"surge_multiplier,omitempty"`
	SurgePrice            *float64        `json:"surge_price,omitempty"`
	CreatedAt             time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// JobAcceptance represents a worker accepting a job
type JobAcceptance struct {
	ID                   string    `json:"id" gorm:"primaryKey"`
	JobID                string    `json:"job_id" gorm:"index;not null"`
	WorkerID             string    `json:"worker_id" gorm:"index;not null"`
	AcceptedAt           time.Time `json:"accepted_at" gorm:"not null"`
	EstimatedArrivalMins int       `json:"estimated_arrival_minutes" gorm:"default:30"`
}

// JobCompletion represents a completed job record
type JobCompletion struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	JobID          string    `json:"job_id" gorm:"uniqueIndex;not null"`
	WorkerID       string    `json:"worker_id" gorm:"index;not null"`
	CompletedAt    time.Time `json:"completed_at" gorm:"not null"`
	ClientRating   *int      `json:"client_rating,omitempty"` // 1-5 (client rates worker)
	ClientReview   *string   `json:"client_review,omitempty"`
	WorkerRating   *int      `json:"worker_rating,omitempty"` // 1-5 (worker rates client)
	WorkerReview   *string   `json:"worker_review,omitempty"`
	WorkerNotes    *string   `json:"worker_notes,omitempty"`
	PhotoProofURLs []string  `json:"photo_proof_urls" gorm:"serializer:json"`
}
