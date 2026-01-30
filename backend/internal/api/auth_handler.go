package api

import (
	"net/http"
	"strings"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// SendOtpRequest is the request body for sending OTP
type SendOtpRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

// VerifyOtpRequest is the request body for verifying OTP
type VerifyOtpRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Otp         string `json:"otp" validate:"required,len=6"`
	RequestID   string `json:"request_id" validate:"required"`
}

// RefreshTokenRequest is the request body for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// SendOtp godoc
// @Summary Send OTP to phone number
// @Description Sends a 6-digit OTP to the provided phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SendOtpRequest true "Phone number"
// @Success 200 {object} service.OtpResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/send-otp [post]
func (h *AuthHandler) SendOtp(c echo.Context) error {
	var req SendOtpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	result, err := h.authService.SendOtp(c.Request().Context(), req.PhoneNumber)
	if err != nil {
		if err == service.ErrInvalidPhoneNumber {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid phone number format"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to send OTP"})
	}

	return c.JSON(http.StatusOK, result)
}

// VerifyOtp godoc
// @Summary Verify OTP and authenticate
// @Description Verifies the OTP and returns auth tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyOtpRequest true "OTP verification details"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOtp(c echo.Context) error {
	var req VerifyOtpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	result, err := h.authService.VerifyOtp(c.Request().Context(), req.PhoneNumber, req.Otp, req.RequestID)
	if err != nil {
		switch err {
		case service.ErrInvalidOtp:
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid OTP"})
		case service.ErrOtpExpired:
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "OTP expired"})
		case service.ErrTooManyAttempts:
			return c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "too many attempts"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "verification failed"})
		}
	}

	return c.JSON(http.StatusOK, result)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Uses refresh token to get new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} domain.AuthTokens
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	tokens, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid refresh token"})
	}

	return c.JSON(http.StatusOK, tokens)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidates all user sessions
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	userID := c.Get("user_id").(string)
	
	if err := h.authService.Logout(c.Request().Context(), userID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "logout failed"})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "logged out successfully"})
}

// Me godoc
// @Summary Get current user
// @Description Returns the authenticated user's information
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	userID := c.Get("user_id").(string)
	phone := c.Get("phone").(string)
	userType := c.Get("user_type").(string)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":   userID,
		"phone":     phone,
		"user_type": userType,
	})
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization format"})
			}

			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			}

			// Set user info in context
			c.Set("user_id", claims.UserID)
			c.Set("phone", claims.Phone)
			c.Set("user_type", string(claims.UserType))

			return next(c)
		}
	}
}
