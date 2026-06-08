package api

import (
	"net/http"
	"strconv"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// LegalHandler handles legal document endpoints
type LegalHandler struct {
	legalService *service.LegalService
}

// NewLegalHandler creates a new legal handler
func NewLegalHandler(legalService *service.LegalService) *LegalHandler {
	return &LegalHandler{legalService: legalService}
}

// GetDocuments returns all active legal documents
// GET /api/v1/legal/documents
func (h *LegalHandler) GetDocuments(c echo.Context) error {
	docs, err := h.legalService.GetDocuments()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch documents"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"documents": docs,
	})
}

// GetDocument returns a single document by slug
// GET /api/v1/legal/documents/:slug
func (h *LegalHandler) GetDocument(c echo.Context) error {
	slug := c.Param("slug")
	doc, err := h.legalService.GetDocumentBySlug(slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "document not found"})
	}
	return c.JSON(http.StatusOK, doc)
}

// AcceptDocumentsRequest is the API request for accepting documents
type AcceptDocumentsAPIRequest struct {
	Acceptances []service.DocumentAcceptance `json:"acceptances"`
	DeviceInfo  string                       `json:"device_info"`
}

// AcceptDocuments records user acceptance of legal documents
// POST /api/v1/legal/accept
func (h *LegalHandler) AcceptDocuments(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req AcceptDocumentsAPIRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if len(req.Acceptances) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no documents specified"})
	}

	serviceReq := service.AcceptDocumentsRequest{
		Acceptances: req.Acceptances,
		IPAddress:   c.RealIP(),
		DeviceInfo:  req.DeviceInfo,
	}

	if err := h.legalService.AcceptDocuments(userID, serviceReq); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "legal documents accepted"})
}

// AcceptWorkerPolicyAPIRequest is the API request for worker policy acceptance
type AcceptWorkerPolicyAPIRequest struct {
	DeviceInfo string `json:"device_info"`
}

// AcceptWorkerPolicy records worker policy acceptance
// POST /api/v1/legal/accept-worker-policy
func (h *LegalHandler) AcceptWorkerPolicy(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req AcceptWorkerPolicyAPIRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	serviceReq := service.AcceptWorkerPolicyRequest{
		IPAddress:  c.RealIP(),
		DeviceInfo: req.DeviceInfo,
	}

	if err := h.legalService.AcceptWorkerPolicy(userID, serviceReq); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "worker policy accepted"})
}

// GetComplianceStatus checks user's legal compliance
// GET /api/v1/legal/compliance
func (h *LegalHandler) GetComplianceStatus(c echo.Context) error {
	userID := c.Get("user_id").(string)
	userType := c.Get("user_type").(string)
	isWorker := userType == "WORKER"

	compliance, err := h.legalService.CheckLegalCompliance(userID, isWorker)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check compliance"})
	}

	return c.JSON(http.StatusOK, compliance)
}

// === Admin Endpoints ===

// UpdateDocumentVersionRequest is the admin request to update a document version
type UpdateDocumentVersionRequest struct {
	Version string `json:"version"`
	PdfHash string `json:"pdf_hash"`
}

// AdminUpdateDocument updates a document version (admin)
// PUT /api/v1/admin/legal/documents/:slug
func (h *LegalHandler) AdminUpdateDocument(c echo.Context) error {
	slug := c.Param("slug")

	var req UpdateDocumentVersionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Version == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "version is required"})
	}

	if err := h.legalService.UpdateDocumentVersion(slug, req.Version, req.PdfHash); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update document"})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "document version updated"})
}

// AdminForceReAccept triggers forced re-acceptance by updating the version
// POST /api/v1/admin/legal/documents/:slug/force-reaccept
func (h *LegalHandler) AdminForceReAccept(c echo.Context) error {
	slug := c.Param("slug")

	doc, err := h.legalService.GetDocumentBySlug(slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "document not found"})
	}

	// Increment version to force re-acceptance
	newVersion := doc.Version + ".1"
	if err := h.legalService.UpdateDocumentVersion(slug, newVersion, doc.PdfHash); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to trigger re-acceptance"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":     "forced re-acceptance triggered",
		"new_version": newVersion,
	})
}

// AdminGetAcceptanceLogs returns paginated acceptance logs
// GET /api/v1/admin/legal/acceptance-logs
func (h *LegalHandler) AdminGetAcceptanceLogs(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	logs, total, err := h.legalService.GetAcceptanceLogs(limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch logs"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

// LegalComplianceMiddleware blocks requests if user hasn't accepted required legal documents
func LegalComplianceMiddleware(legalService *service.LegalService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(string)
			if !ok || userID == "" {
				return next(c) // Let auth middleware handle this
			}

			userType, _ := c.Get("user_type").(string)
			isWorker := userType == "WORKER"

			compliance, err := legalService.CheckLegalCompliance(userID, isWorker)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to verify legal compliance"})
			}

			if !compliance.IsCompliant {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error":      "legal_acceptance_required",
					"message":    "You must accept all legal documents before proceeding",
					"compliance": compliance,
				})
			}

			return next(c)
		}
	}
}
