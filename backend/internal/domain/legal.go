package domain

import "time"

// LegalDocument represents a versioned legal document
type LegalDocument struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Slug          string    `json:"slug" gorm:"uniqueIndex;not null"`
	Title         string    `json:"title" gorm:"not null"`
	Version       string    `json:"version" gorm:"not null;default:v1.0"`
	EffectiveDate time.Time `json:"effective_date" gorm:"not null"`
	PdfHash       string    `json:"pdf_hash" gorm:"not null;default:''"`
	IsActive      bool      `json:"is_active" gorm:"not null;default:true"`
	WorkerOnly    bool      `json:"worker_only" gorm:"not null;default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserLegalAcceptance records when a user accepts a legal document
type UserLegalAcceptance struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	UserID          string    `json:"user_id" gorm:"not null"`
	DocumentID      string    `json:"document_id" gorm:"not null"`
	DocumentSlug    string    `json:"document_slug" gorm:"not null"`
	AcceptedVersion string    `json:"accepted_version" gorm:"not null"`
	AcceptedAt      time.Time `json:"accepted_at" gorm:"not null"`
	IPAddress       string    `json:"ip_address" gorm:"not null;default:''"`
	DeviceInfo      string    `json:"device_info" gorm:"not null;default:''"`
	PdfHash         string    `json:"pdf_hash" gorm:"not null;default:''"`
}

// WorkerPolicyAcceptance records worker-specific policy acceptance
type WorkerPolicyAcceptance struct {
	ID                  string    `json:"id" gorm:"primaryKey"`
	UserID              string    `json:"user_id" gorm:"uniqueIndex;not null"`
	WorkerPolicyVersion string    `json:"worker_policy_version" gorm:"not null"`
	AcceptedAt          time.Time `json:"accepted_at" gorm:"not null"`
	IPAddress           string    `json:"ip_address" gorm:"not null;default:''"`
	DeviceInfo          string    `json:"device_info" gorm:"not null;default:''"`
	PdfHash             string    `json:"pdf_hash" gorm:"not null;default:''"`
}

// LegalAcceptanceStatus represents the acceptance state for a user
type LegalAcceptanceStatus struct {
	DocumentSlug    string `json:"document_slug"`
	DocumentTitle   string `json:"document_title"`
	CurrentVersion  string `json:"current_version"`
	AcceptedVersion string `json:"accepted_version,omitempty"`
	IsAccepted      bool   `json:"is_accepted"`
	NeedsReAccept   bool   `json:"needs_re_accept"`
}

// LegalComplianceResponse is returned by the compliance endpoint
type LegalComplianceResponse struct {
	IsCompliant          bool                    `json:"is_compliant"`
	WorkerPolicyAccepted bool                    `json:"worker_policy_accepted,omitempty"`
	Documents            []LegalAcceptanceStatus `json:"documents"`
}
