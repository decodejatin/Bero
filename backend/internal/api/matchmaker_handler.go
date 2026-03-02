package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/matchmaker"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// MatchmakerHandler handles matchmaker-related HTTP endpoints.
type MatchmakerHandler struct {
	matchmakerService service.MatchmakerService
}

// NewMatchmakerHandler creates a new matchmaker handler.
func NewMatchmakerHandler(ms service.MatchmakerService) *MatchmakerHandler {
	return &MatchmakerHandler{matchmakerService: ms}
}

// TriggerMatching manually triggers a matching round.
// POST /api/v1/matchmaker/trigger
func (h *MatchmakerHandler) TriggerMatching(c echo.Context) error {
	h.matchmakerService.Trigger()
	return c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Matching round triggered",
	})
}

// GetConfig returns the current matching configuration.
// GET /api/v1/matchmaker/config
func (h *MatchmakerHandler) GetConfig(c echo.Context) error {
	cfg := h.matchmakerService.GetConfig()
	return c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Current matching configuration",
		Data:    cfg,
	})
}

// UpdateConfig updates the matching configuration.
// PUT /api/v1/matchmaker/config
func (h *MatchmakerHandler) UpdateConfig(c echo.Context) error {
	var cfg matchmaker.MatchConfig
	if err := c.Bind(&cfg); err != nil {
		return c.JSON(http.StatusBadRequest, SuccessResponse{
			Success: false,
			Message: "Invalid configuration: " + err.Error(),
		})
	}

	// Validate weights
	if cfg.AlphaProximity < 0 || cfg.AlphaReputation < 0 || cfg.AlphaSkillMatch < 0 || cfg.AlphaWaitTime < 0 {
		return c.JSON(http.StatusBadRequest, SuccessResponse{
			Success: false,
			Message: "Weight coefficients must be non-negative",
		})
	}

	if cfg.WindowDurationSeconds < 5 {
		return c.JSON(http.StatusBadRequest, SuccessResponse{
			Success: false,
			Message: "Window duration must be at least 5 seconds",
		})
	}

	if cfg.MaxDistanceKm <= 0 {
		return c.JSON(http.StatusBadRequest, SuccessResponse{
			Success: false,
			Message: "Max distance must be positive",
		})
	}

	h.matchmakerService.UpdateConfig(cfg)
	return c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Configuration updated",
		Data:    cfg,
	})
}

// GetStatus returns the current engine status.
// GET /api/v1/matchmaker/status
func (h *MatchmakerHandler) GetStatus(c echo.Context) error {
	status := h.matchmakerService.GetStatus()
	return c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Engine status",
		Data:    status,
	})
}
