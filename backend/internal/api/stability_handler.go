package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// StabilityHandler handles stability enforcement API endpoints.
type StabilityHandler struct {
	stabilityService service.StabilityService
}

// NewStabilityHandler creates a new stability handler.
func NewStabilityHandler(stabilityService service.StabilityService) *StabilityHandler {
	return &StabilityHandler{stabilityService: stabilityService}
}

// --- Request types ---

// UpdateStabilityConfigRequest for PUT /stability/config
type UpdateStabilityConfigRequest struct {
	SwitchCostFixed        *float64 `json:"switch_cost_fixed,omitempty"`
	EarningsPenaltyPercent *float64 `json:"earnings_penalty_percent,omitempty"`
	MaxCancelsPerHour      *int     `json:"max_cancels_per_hour,omitempty"`
	EscalationThreshold    *int     `json:"escalation_threshold,omitempty"`
	CooldownMinutes        *int     `json:"cooldown_minutes,omitempty"`
	DecayLambda            *float64 `json:"decay_lambda,omitempty"`
	TravelCostPerKm        *float64 `json:"travel_cost_per_km,omitempty"`
	VisibilityPenalty      *float64 `json:"visibility_penalty,omitempty"`
}

// ReassignRequest for POST /stability/reassign
type ReassignRequest struct {
	WorkerID     string `json:"worker_id"`
	CurrentJobID string `json:"current_job_id"`
	NewJobID     string `json:"new_job_id"`
}

// --- Handlers ---

// GetStats handles GET /api/v1/stability/stats
func (h *StabilityHandler) GetStats(c echo.Context) error {
	stats, err := h.stabilityService.GetStabilityStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load stats"})
	}
	return c.JSON(http.StatusOK, stats)
}

// GetConfig handles GET /api/v1/stability/config
func (h *StabilityHandler) GetConfig(c echo.Context) error {
	cfg, err := h.stabilityService.GetConfig(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load config"})
	}
	return c.JSON(http.StatusOK, cfg)
}

// UpdateConfig handles PUT /api/v1/stability/config
func (h *StabilityHandler) UpdateConfig(c echo.Context) error {
	var req UpdateStabilityConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Load current config and apply partial updates
	current, err := h.stabilityService.GetConfig(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load current config"})
	}

	if req.SwitchCostFixed != nil {
		current.SwitchCostFixed = *req.SwitchCostFixed
	}
	if req.EarningsPenaltyPercent != nil {
		current.EarningsPenaltyPercent = *req.EarningsPenaltyPercent
	}
	if req.MaxCancelsPerHour != nil {
		current.MaxCancelsPerHour = *req.MaxCancelsPerHour
	}
	if req.EscalationThreshold != nil {
		current.EscalationThreshold = *req.EscalationThreshold
	}
	if req.CooldownMinutes != nil {
		current.CooldownMinutes = *req.CooldownMinutes
	}
	if req.DecayLambda != nil {
		current.DecayLambda = *req.DecayLambda
	}
	if req.TravelCostPerKm != nil {
		current.TravelCostPerKm = *req.TravelCostPerKm
	}
	if req.VisibilityPenalty != nil {
		current.VisibilityPenalty = *req.VisibilityPenalty
	}

	if err := h.stabilityService.UpdateConfig(c.Request().Context(), current); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update config"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
		"config": current,
	})
}

// GetUserStatus handles GET /api/v1/stability/user/:id/status
func (h *StabilityHandler) GetUserStatus(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user id required"})
	}

	status, err := h.stabilityService.GetUserStatus(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, status)
}

// CheckReassign handles POST /api/v1/stability/reassign
// Checks whether a worker can switch from current job to new job.
func (h *StabilityHandler) CheckReassign(c echo.Context) error {
	var req ReassignRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.WorkerID == "" || req.CurrentJobID == "" || req.NewJobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "worker_id, current_job_id, and new_job_id are required"})
	}

	allowed, reason, err := h.stabilityService.CanReassign(
		c.Request().Context(), req.WorkerID, req.CurrentJobID, req.NewJobID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	status := "denied"
	httpStatus := http.StatusForbidden
	if allowed {
		status = "allowed"
		httpStatus = http.StatusOK
	}

	return c.JSON(httpStatus, map[string]interface{}{
		"status":  status,
		"allowed": allowed,
		"reason":  reason,
	})
}

// ComputeUtility handles POST /api/v1/stability/utility
// Computes the time-decaying utility for a worker-job pair.
func (h *StabilityHandler) ComputeUtility(c echo.Context) error {
	var req struct {
		JobValueRupees float64 `json:"job_value_rupees"`
		DistanceKm     float64 `json:"distance_km"`
		RatingAvg      float64 `json:"rating_avg"`
		DelayMinutes   float64 `json:"delay_minutes"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	utility := h.stabilityService.ComputeUtility(
		req.JobValueRupees, req.DistanceKm, req.RatingAvg, req.DelayMinutes,
	)

	return c.JSON(http.StatusOK, utility)
}

// Ensure ErrorResponse and SuccessResponse are used from existing types
var _ = domain.StabilityConfig{}
