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
