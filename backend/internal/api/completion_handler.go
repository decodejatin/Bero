package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// CompletionHandler handles dual-sided completion and rating endpoints.
type CompletionHandler struct {
	completionService service.CompletionService
}

// NewCompletionHandler creates a new completion handler.
func NewCompletionHandler(completionService service.CompletionService) *CompletionHandler {
	return &CompletionHandler{completionService: completionService}
}

// WorkerComplete handles POST /api/v1/jobs/:id/complete-by-worker
func (h *CompletionHandler) WorkerComplete(c echo.Context) error {
	jobID := c.Param("id")
	var req struct {
		WorkerID string `json:"worker_id"`
	}
	if err := c.Bind(&req); err != nil || req.WorkerID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "worker_id required"})
	}

	if err := h.completionService.WorkerMarkComplete(c.Request().Context(), jobID, req.WorkerID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	status, _ := h.completionService.GetCompletionStatus(c.Request().Context(), jobID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "worker_completed",
		"details": status,
	})
}

// ClientConfirm handles POST /api/v1/jobs/:id/confirm-by-client
func (h *CompletionHandler) ClientConfirm(c echo.Context) error {
	jobID := c.Param("id")
	var req struct {
		ClientID string `json:"client_id"`
	}
	if err := c.Bind(&req); err != nil || req.ClientID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "client_id required"})
	}

	if err := h.completionService.ClientConfirmComplete(c.Request().Context(), jobID, req.ClientID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	status, _ := h.completionService.GetCompletionStatus(c.Request().Context(), jobID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "client_confirmed",
		"details": status,
	})
}

// SubmitRating handles POST /api/v1/jobs/:id/rate
func (h *CompletionHandler) SubmitRating(c echo.Context) error {
	jobID := c.Param("id")
	var req struct {
		RaterID string `json:"rater_id"`
		Rating  int    `json:"rating"`
		Review  string `json:"review"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if req.RaterID == "" || req.Rating == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "rater_id and rating required"})
	}

	if err := h.completionService.SubmitMutualRating(c.Request().Context(), jobID, req.RaterID, req.Rating, req.Review); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	status, _ := h.completionService.GetCompletionStatus(c.Request().Context(), jobID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "rating_submitted",
		"details": status,
	})
}

// GetCompletionStatus handles GET /api/v1/jobs/:id/completion-status
func (h *CompletionHandler) GetCompletionStatus(c echo.Context) error {
	jobID := c.Param("id")
	status, err := h.completionService.GetCompletionStatus(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, status)
}

// CheckWorkerBlocked handles GET /api/v1/completion/blocked/worker/:id
func (h *CompletionHandler) CheckWorkerBlocked(c echo.Context) error {
	workerID := c.Param("id")
	blocked, reason, err := h.completionService.IsWorkerBlockedByRating(c.Request().Context(), workerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"blocked": blocked,
		"reason":  reason,
	})
}

// CheckClientBlocked handles GET /api/v1/completion/blocked/client/:id
func (h *CompletionHandler) CheckClientBlocked(c echo.Context) error {
	clientID := c.Param("id")
	blocked, reason, err := h.completionService.IsClientBlockedByRating(c.Request().Context(), clientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"blocked": blocked,
		"reason":  reason,
	})
}
