package api

import (
	"net/http"
	"strconv"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// PricingHandler handles dynamic pricing API endpoints.
type PricingHandler struct {
	pricingService service.PricingService
}

// NewPricingHandler creates a new pricing handler.
func NewPricingHandler(pricingService service.PricingService) *PricingHandler {
	return &PricingHandler{pricingService: pricingService}
}

// GetSurge handles GET /api/v1/pricing/surge?h3=...
// Returns current surge multiplier for a specific H3 hexagon.
func (h *PricingHandler) GetSurge(c echo.Context) error {
	h3Index := c.QueryParam("h3")
	if h3Index == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "h3 query param required"})
	}

	surge, err := h.pricingService.ComputeSurge(c.Request().Context(), h3Index)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, surge)
}

// GetPriceQuote handles GET /api/v1/pricing/quote/:jobId
// Returns full price quote with surge for a job.
func (h *PricingHandler) GetPriceQuote(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "jobId required"})
	}

	quote, err := h.pricingService.GetPriceQuote(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, quote)
}

// ApplySurge handles POST /api/v1/pricing/apply/:jobId
// Applies surge pricing to a job and stores snapshot.
func (h *PricingHandler) ApplySurge(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "jobId required"})
	}

	quote, err := h.pricingService.ApplySurgeToJob(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, quote)
}

// GetConfig handles GET /api/v1/pricing/config
func (h *PricingHandler) GetConfig(c echo.Context) error {
	cfg, err := h.pricingService.GetConfig(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load config"})
	}
	return c.JSON(http.StatusOK, cfg)
}

// UpdateConfig handles PUT /api/v1/pricing/config
func (h *PricingHandler) UpdateConfig(c echo.Context) error {
	var req struct {
		MaxSurgeMultiplier    *float64 `json:"max_surge_multiplier,omitempty"`
		EquilibriumTheta      *float64 `json:"equilibrium_theta,omitempty"`
		ElasticitySensitivity *float64 `json:"elasticity_sensitivity,omitempty"`
		EmergencySurgeCap     *float64 `json:"emergency_surge_cap,omitempty"`
		MaxSurgeChangeRate    *float64 `json:"max_surge_change_rate,omitempty"`
		MinSurgeMultiplier    *float64 `json:"min_surge_multiplier,omitempty"`
		SupplyKRingRadius     *int     `json:"supply_kring_radius,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	current, err := h.pricingService.GetConfig(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load config"})
	}

	if req.MaxSurgeMultiplier != nil {
		current.MaxSurgeMultiplier = *req.MaxSurgeMultiplier
	}
	if req.EquilibriumTheta != nil {
		current.EquilibriumTheta = *req.EquilibriumTheta
	}
	if req.ElasticitySensitivity != nil {
		current.ElasticitySensitivity = *req.ElasticitySensitivity
	}
	if req.EmergencySurgeCap != nil {
		current.EmergencySurgeCap = *req.EmergencySurgeCap
	}
	if req.MaxSurgeChangeRate != nil {
		current.MaxSurgeChangeRate = *req.MaxSurgeChangeRate
	}
	if req.MinSurgeMultiplier != nil {
		current.MinSurgeMultiplier = *req.MinSurgeMultiplier
	}
	if req.SupplyKRingRadius != nil {
		current.SupplyKRingRadius = *req.SupplyKRingRadius
	}

	if err := h.pricingService.UpdateConfig(c.Request().Context(), current); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update config"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
		"config": current,
	})
}

// GetSurgeHistory handles GET /api/v1/pricing/history?h3=...&limit=...
func (h *PricingHandler) GetSurgeHistory(c echo.Context) error {
	h3Index := c.QueryParam("h3")
	if h3Index == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "h3 query param required"})
	}

	limit := 50
	if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
		limit = l
	}

	history, err := h.pricingService.GetSurgeHistory(c.Request().Context(), h3Index, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"h3_index": h3Index,
		"count":    len(history),
		"history":  history,
	})
}
