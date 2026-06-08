package repository

import (
	"github.com/decodejatin/bero-backend/internal/domain"
	"gorm.io/gorm"
)

// LegalRepository handles legal document persistence
type LegalRepository struct {
	db *gorm.DB
}

// NewLegalRepository creates a new legal repository
func NewLegalRepository(db *gorm.DB) *LegalRepository {
	return &LegalRepository{db: db}
}

// GetActiveDocuments returns all active legal documents
func (r *LegalRepository) GetActiveDocuments() ([]domain.LegalDocument, error) {
	var docs []domain.LegalDocument
	err := r.db.Where("is_active = ?", true).Order("worker_only ASC, title ASC").Find(&docs).Error
	return docs, err
}

// GetDocumentBySlug returns a single document by slug
func (r *LegalRepository) GetDocumentBySlug(slug string) (*domain.LegalDocument, error) {
	var doc domain.LegalDocument
	err := r.db.Where("slug = ? AND is_active = ?", slug, true).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// RecordAcceptance inserts a user legal acceptance record
func (r *LegalRepository) RecordAcceptance(acceptance *domain.UserLegalAcceptance) error {
	return r.db.Create(acceptance).Error
}

// GetUserAcceptances returns all acceptance records for a user
func (r *LegalRepository) GetUserAcceptances(userID string) ([]domain.UserLegalAcceptance, error) {
	var acceptances []domain.UserLegalAcceptance
	// Get the latest acceptance per document slug for the user
	err := r.db.Raw(`
		SELECT DISTINCT ON (document_slug) *
		FROM user_legal_acceptance
		WHERE user_id = ?
		ORDER BY document_slug, accepted_at DESC
	`, userID).Scan(&acceptances).Error
	return acceptances, err
}

// RecordWorkerPolicyAcceptance inserts or updates worker policy acceptance
func (r *LegalRepository) RecordWorkerPolicyAcceptance(acceptance *domain.WorkerPolicyAcceptance) error {
	return r.db.Save(acceptance).Error
}

// GetWorkerPolicyAcceptance gets the worker policy acceptance for a user
func (r *LegalRepository) GetWorkerPolicyAcceptance(userID string) (*domain.WorkerPolicyAcceptance, error) {
	var acceptance domain.WorkerPolicyAcceptance
	err := r.db.Where("user_id = ?", userID).First(&acceptance).Error
	if err != nil {
		return nil, err
	}
	return &acceptance, nil
}

// UpdateDocumentVersion updates a document's version and effective date
func (r *LegalRepository) UpdateDocumentVersion(slug, version, pdfHash string) error {
	return r.db.Model(&domain.LegalDocument{}).
		Where("slug = ?", slug).
		Updates(map[string]interface{}{
			"version":        version,
			"pdf_hash":       pdfHash,
			"effective_date": gorm.Expr("NOW()"),
			"updated_at":     gorm.Expr("NOW()"),
		}).Error
}

// GetAcceptanceLogs returns paginated acceptance logs for admin
func (r *LegalRepository) GetAcceptanceLogs(limit, offset int) ([]domain.UserLegalAcceptance, int64, error) {
	var total int64
	r.db.Model(&domain.UserLegalAcceptance{}).Count(&total)

	var logs []domain.UserLegalAcceptance
	err := r.db.Order("accepted_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}
