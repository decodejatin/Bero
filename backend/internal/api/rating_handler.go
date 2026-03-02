package api

import (
	"net/http"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// RatingHandler handles rating endpoints
type RatingHandler struct {
	ratingService service.RatingService
}

func NewRatingHandler(ratingService service.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

type SubmitRatingRequest struct {
	Rating int      `json:"rating"`
	Review string   `json:"review"`
	Tags   []string `json:"tags"`
}

// SubmitRating submits a rating for a job
func (h *RatingHandler) SubmitRating(c echo.Context) error {
	userID := c.Get("user_id").(string)
	jobID := c.Param("id")

	var req SubmitRatingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Rating < 1 || req.Rating > 5 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "rating must be between 1 and 5"})
	}

	err := h.ratingService.SubmitRating(c.Request().Context(), jobID, userID, req.Rating, req.Review, req.Tags)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "rating submitted successfully"})
}

// GetJobRatings returns ratings for a job
func (h *RatingHandler) GetJobRatings(c echo.Context) error {
	jobID := c.Param("id")

	ratings, err := h.ratingService.GetJobRatings(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, ratings)
}
