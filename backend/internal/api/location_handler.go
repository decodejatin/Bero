package api

import (
	"net/http"
	"strconv"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// LocationHandler handles geospatial endpoints
type LocationHandler struct {
	locationService service.LocationService
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{locationService: locationService}
}

// --- Request / Response types ---

// UpdateLocationRequest is the body for PUT /workers/location
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// UpdateLocationResponse is returned after a location update
type UpdateLocationResponse struct {
	Status  string `json:"status"`
	H3Index string `json:"h3_index"`
}

// SetAvailabilityRequest is the body for PUT /workers/availability
type SetAvailabilityRequest struct {
	IsAvailable bool `json:"is_available"`
}

// --- Handlers ---

// UpdateWorkerLocation handles PUT /api/v1/workers/location
// Called every 10 seconds by the mobile location service.
func (h *LocationHandler) UpdateWorkerLocation(c echo.Context) error {
	workerID := c.Get("user_id").(string)

	var req UpdateLocationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Latitude == 0 && req.Longitude == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "latitude and longitude are required"})
	}

	h3Index, err := h.locationService.UpdateWorkerLocation(c.Request().Context(), workerID, req.Latitude, req.Longitude)
	if err != nil {
		if err == service.ErrInvalidCoordinates {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update location"})
	}

	return c.JSON(http.StatusOK, UpdateLocationResponse{
		Status:  "ok",
		H3Index: h3Index,
	})
}

// GetNearbyWorkers handles GET /api/v1/workers/nearby?lat=...&lon=...&radius=...
// Returns available workers ordered by distance ASC, limited to 20.
// This endpoint powers the candidate list for Hungarian matching.
func (h *LocationHandler) GetNearbyWorkers(c echo.Context) error {
	latStr := c.QueryParam("lat")
	lonStr := c.QueryParam("lon")
	radiusStr := c.QueryParam("radius")

	if latStr == "" || lonStr == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "lat and lon query parameters are required"})
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lat parameter"})
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lon parameter"})
	}

	var radius float64
	if radiusStr != "" {
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid radius parameter"})
		}
	}

	result, err := h.locationService.GetNearbyWorkers(c.Request().Context(), lat, lon, radius)
	if err != nil {
		if err == service.ErrInvalidCoordinates {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to query nearby workers"})
	}

	return c.JSON(http.StatusOK, result)
}

// SetAvailability handles PUT /api/v1/workers/availability
func (h *LocationHandler) SetAvailability(c echo.Context) error {
	workerID := c.Get("user_id").(string)

	var req SetAvailabilityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if err := h.locationService.SetWorkerAvailability(c.Request().Context(), workerID, req.IsAvailable); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update availability"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       "ok",
		"is_available": req.IsAvailable,
	})
}

// GoOffline handles DELETE /api/v1/workers/location
// Worker goes offline — sets is_available to false.
func (h *LocationHandler) GoOffline(c echo.Context) error {
	workerID := c.Get("user_id").(string)

	if err := h.locationService.SetWorkerOffline(c.Request().Context(), workerID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to go offline"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "offline",
	})
}
