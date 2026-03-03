package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// ProfileHandler handles profile endpoints
type ProfileHandler struct {
	profileService service.ProfileService
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(profileService service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// UpdateProfileRequest is the request body for updating profile
type UpdateProfileRequest struct {
	FullName string  `json:"full_name"`
	Email    *string `json:"email,omitempty"`
	Address  *string `json:"address,omitempty"`
}

// SetUserTypeRequest is the request body for setting user type
type SetUserTypeRequest struct {
	UserType string `json:"user_type" validate:"required,oneof=WORKER CLIENT"`
}

// GetProfile godoc
// @Summary Get current user's profile
// @Description Returns the authenticated user's complete profile
// @Tags profile
// @Security BearerAuth
// @Success 200 {object} service.ProfileResponse
// @Failure 401 {object} ErrorResponse
// @Router /profile [get]
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	profile, err := h.profileService.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get profile"})
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Update user's profile
// @Description Updates the authenticated user's profile information
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile data"
// @Success 200 {object} service.ProfileResponse
// @Failure 400 {object} ErrorResponse
// @Router /profile [put]
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	profile, err := h.profileService.UpdateProfile(c.Request().Context(), userID, req.FullName, req.Email, req.Address)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile"})
	}

	return c.JSON(http.StatusOK, profile)
}

// SetUserType godoc
// @Summary Set user type (role)
// @Description Sets the user type after role selection (WORKER or CLIENT)
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SetUserTypeRequest true "User type"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /profile/user-type [post]
func (h *ProfileHandler) SetUserType(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req SetUserTypeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.UserType != "WORKER" && req.UserType != "CLIENT" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user_type must be WORKER or CLIENT"})
	}

	if err := h.profileService.SetUserType(c.Request().Context(), userID, req.UserType); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to set user type"})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "user type set successfully"})
}

// GetProfileById godoc
// @Summary Get a user's public profile by ID
// @Description Returns a user's public profile information for viewing worker details
// @Tags profile
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} service.ProfileResponse
// @Failure 404 {object} ErrorResponse
// @Router /profile/{id} [get]
func (h *ProfileHandler) GetProfileById(c echo.Context) error {
	profileID := c.Param("id")

	if profileID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user id is required"})
	}

	profile, err := h.profileService.GetProfile(c.Request().Context(), profileID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile not found"})
	}

	return c.JSON(http.StatusOK, profile)
}

// GetUserStats godoc
// @Summary Get user stats
// @Description Returns stats for the authenticated user (jobs posted/completed, total spent/earned, avg rating)
// @Tags profile
// @Security BearerAuth
// @Success 200 {object} service.UserStatsResponse
// @Failure 401 {object} ErrorResponse
// @Router /profile/stats [get]
func (h *ProfileHandler) GetUserStats(c echo.Context) error {
	userID := c.Get("user_id").(string)

	stats, err := h.profileService.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get stats"})
	}

	return c.JSON(http.StatusOK, stats)
}

// UpdateWorkerSkillsRequest is the request body for updating worker skills
type UpdateWorkerSkillsRequest struct {
	Skills []string `json:"skills"`
}

// UpdateWorkerSkills godoc
// @Summary Update worker skills
// @Description Updates the authenticated worker's skills list
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateWorkerSkillsRequest true "Skills list"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /profile/worker/skills [put]
func (h *ProfileHandler) UpdateWorkerSkills(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req UpdateWorkerSkillsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if err := h.profileService.UpdateWorkerSkills(c.Request().Context(), userID, req.Skills); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update skills: " + err.Error()})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "skills updated successfully"})
}

// GetMyRatings godoc
// @Summary Get my rating history
// @Description Returns all ratings given and received by the authenticated user
// @Tags profile
// @Security BearerAuth
// @Success 200 {array} service.RatingHistoryItem
// @Failure 401 {object} ErrorResponse
// @Router /profile/ratings [get]
func (h *ProfileHandler) GetMyRatings(c echo.Context) error {
	userID := c.Get("user_id").(string)

	ratings, err := h.profileService.GetMyRatings(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get ratings"})
	}

	return c.JSON(http.StatusOK, ratings)
}
