package service

import (
	"fmt"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/google/uuid"
)

// LegalService handles legal document business logic
type LegalService struct {
	legalRepo *repository.LegalRepository
}

// NewLegalService creates a new legal service
func NewLegalService(legalRepo *repository.LegalRepository) *LegalService {
	return &LegalService{legalRepo: legalRepo}
}

// GetDocuments returns all active legal documents
func (s *LegalService) GetDocuments() ([]domain.LegalDocument, error) {
	return s.legalRepo.GetActiveDocuments()
}

// GetDocumentBySlug returns a single document
func (s *LegalService) GetDocumentBySlug(slug string) (*domain.LegalDocument, error) {
	return s.legalRepo.GetDocumentBySlug(slug)
}

// AcceptDocumentsRequest is the request body for accepting documents
type AcceptDocumentsRequest struct {
	Acceptances []DocumentAcceptance `json:"acceptances"`
	IPAddress   string               `json:"ip_address"`
	DeviceInfo  string               `json:"device_info"`
}

// DocumentAcceptance represents acceptance of a single document
type DocumentAcceptance struct {
	DocumentSlug string `json:"document_slug"`
	Version      string `json:"version"`
}

// AcceptDocuments records the user's acceptance of legal documents
func (s *LegalService) AcceptDocuments(userID string, req AcceptDocumentsRequest) error {
	for _, acc := range req.Acceptances {
		doc, err := s.legalRepo.GetDocumentBySlug(acc.DocumentSlug)
		if err != nil {
			return fmt.Errorf("document not found: %s", acc.DocumentSlug)
		}

		acceptance := &domain.UserLegalAcceptance{
			ID:              uuid.New().String(),
			UserID:          userID,
			DocumentID:      doc.ID,
			DocumentSlug:    doc.Slug,
			AcceptedVersion: doc.Version,
			AcceptedAt:      time.Now(),
			IPAddress:       req.IPAddress,
			DeviceInfo:      req.DeviceInfo,
			PdfHash:         doc.PdfHash,
		}

		if err := s.legalRepo.RecordAcceptance(acceptance); err != nil {
			return fmt.Errorf("failed to record acceptance for %s: %w", acc.DocumentSlug, err)
		}
	}
	return nil
}

// AcceptWorkerPolicyRequest is the request body for worker policy acceptance
type AcceptWorkerPolicyRequest struct {
	Version    string `json:"version"`
	IPAddress  string `json:"ip_address"`
	DeviceInfo string `json:"device_info"`
}

// AcceptWorkerPolicy records the worker's policy acceptance
func (s *LegalService) AcceptWorkerPolicy(userID string, req AcceptWorkerPolicyRequest) error {
	doc, err := s.legalRepo.GetDocumentBySlug("worker-responsibility")
	if err != nil {
		return fmt.Errorf("worker responsibility document not found")
	}

	acceptance := &domain.WorkerPolicyAcceptance{
		ID:                  uuid.New().String(),
		UserID:              userID,
		WorkerPolicyVersion: doc.Version,
		AcceptedAt:          time.Now(),
		IPAddress:           req.IPAddress,
		DeviceInfo:          req.DeviceInfo,
		PdfHash:             doc.PdfHash,
	}

	return s.legalRepo.RecordWorkerPolicyAcceptance(acceptance)
}

// CheckLegalCompliance checks whether a user has accepted all required documents at current versions
func (s *LegalService) CheckLegalCompliance(userID string, isWorker bool) (*domain.LegalComplianceResponse, error) {
	docs, err := s.legalRepo.GetActiveDocuments()
	if err != nil {
		return nil, err
	}

	acceptances, err := s.legalRepo.GetUserAcceptances(userID)
	if err != nil {
		return nil, err
	}

	// Build acceptance map: slug -> latest acceptance
	acceptMap := make(map[string]*domain.UserLegalAcceptance)
	for i := range acceptances {
		acceptMap[acceptances[i].DocumentSlug] = &acceptances[i]
	}

	isCompliant := true
	var statuses []domain.LegalAcceptanceStatus

	for _, doc := range docs {
		// Skip worker-only docs for non-workers
		if doc.WorkerOnly && !isWorker {
			continue
		}

		status := domain.LegalAcceptanceStatus{
			DocumentSlug:   doc.Slug,
			DocumentTitle:  doc.Title,
			CurrentVersion: doc.Version,
		}

		if acc, ok := acceptMap[doc.Slug]; ok {
			status.AcceptedVersion = acc.AcceptedVersion
			status.IsAccepted = true
			// Check if version is outdated
			if acc.AcceptedVersion != doc.Version {
				status.NeedsReAccept = true
				status.IsAccepted = false
				isCompliant = false
			}
		} else {
			isCompliant = false
		}

		statuses = append(statuses, status)
	}

	response := &domain.LegalComplianceResponse{
		IsCompliant: isCompliant,
		Documents:   statuses,
	}

	// Check worker policy if the user is a worker
	if isWorker {
		wpAcceptance, wpErr := s.legalRepo.GetWorkerPolicyAcceptance(userID)
		if wpErr != nil || wpAcceptance == nil {
			response.WorkerPolicyAccepted = false
			response.IsCompliant = false
		} else {
			// Verify against current worker-responsibility version
			workerDoc, _ := s.legalRepo.GetDocumentBySlug("worker-responsibility")
			if workerDoc != nil && wpAcceptance.WorkerPolicyVersion != workerDoc.Version {
				response.WorkerPolicyAccepted = false
				response.IsCompliant = false
			} else {
				response.WorkerPolicyAccepted = true
			}
		}
	}

	return response, nil
}

// UpdateDocumentVersion updates a document version (admin)
func (s *LegalService) UpdateDocumentVersion(slug, version, pdfHash string) error {
	return s.legalRepo.UpdateDocumentVersion(slug, version, pdfHash)
}

// GetAcceptanceLogs returns paginated acceptance logs (admin)
func (s *LegalService) GetAcceptanceLogs(limit, offset int) ([]domain.UserLegalAcceptance, int64, error) {
	return s.legalRepo.GetAcceptanceLogs(limit, offset)
}
