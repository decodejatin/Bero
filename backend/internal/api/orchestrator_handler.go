package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

// OrchestratorHandler exposes the unified dispatch pipeline via REST.
type OrchestratorHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewOrchestratorHandler creates a new orchestrator handler.
func NewOrchestratorHandler(orch *orchestrator.Orchestrator) *OrchestratorHandler {
	return &OrchestratorHandler{orchestrator: orch}
}

// SubmitJob handles POST /api/v1/pipeline/submit
// Single entry point: applies pricing + enqueues for matching.
func (h *OrchestratorHandler) SubmitJob(c echo.Context) error {
	var req struct {
		JobID string `json:"job_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if req.JobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job_id required"})
	}

	quote, err := h.orchestrator.SubmitJob(c.Request().Context(), req.JobID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	resp := map[string]interface{}{
		"status": "submitted",
		"job_id": req.JobID,
	}
	if quote != nil {
		resp["surge_multiplier"] = quote.SurgeMultiplier
		resp["final_price"] = quote.FinalPrice
	}

	return c.JSON(http.StatusAccepted, resp)
}

// GetStatus handles GET /api/v1/pipeline/status
// Returns pipeline status, queue depth, last batch result, lifetime stats.
func (h *OrchestratorHandler) GetStatus(c echo.Context) error {
	status := h.orchestrator.GetStatus()
	return c.JSON(http.StatusOK, status)
}
