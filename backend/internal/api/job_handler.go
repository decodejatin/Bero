package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// JobHandler handles job endpoints
type JobHandler struct {
	jobService service.JobService
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobService service.JobService) *JobHandler {
	return &JobHandler{jobService: jobService}
}

// CreateJobRequest is the request body for creating a job
type CreateJobRequest struct {
	Title                 string                 `json:"title"`
	Description           string                 `json:"description"`
	Category              domain.ServiceCategory `json:"category"`
	ClientName            string                 `json:"client_name"`
	Address               string                 `json:"address"`
	Locality              string                 `json:"locality"`
	City                  string                 `json:"city"`
	Pincode               string                 `json:"pincode"`
	Latitude              *float64               `json:"latitude,omitempty"`
	Longitude             *float64               `json:"longitude,omitempty"`
	EstimatedDurationMins int                    `json:"estimated_duration_minutes"`
	PaymentAmountRupees   float64                `json:"payment_amount_rupees"`
	ScheduledDate         string                 `json:"scheduled_date"`
	ScheduledTimeSlot     string                 `json:"scheduled_time_slot"`
	IsUrgent              bool                   `json:"is_urgent"`
	RequiredSkills        []string               `json:"required_skills"`
}

// CompleteJobRequest is the request body for completing a job
type CompleteJobRequest struct {
	Notes     *string  `json:"notes,omitempty"`
	PhotoURLs []string `json:"photo_urls,omitempty"`
}

// parseDate parses a date string — supports RFC3339, ISO date-time, and YYYY-MM-DD
func parseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 first (e.g. 2026-02-25T00:00:00Z)
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	// Try RFC3339Nano
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t, nil
	}
	// Try plain date
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}
	// Try date-time without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", dateStr); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// CreateJob handles POST /api/v1/jobs
func (h *JobHandler) CreateJob(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req CreateJobRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Parse scheduled date
	var scheduledDate time.Time
	if req.ScheduledDate != "" {
		t, err := parseDate(req.ScheduledDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid scheduled_date format, use RFC3339 or YYYY-MM-DD"})
		}
		scheduledDate = t
	} else {
		scheduledDate = time.Now()
	}

	job := &domain.Job{
		Title:                 req.Title,
		Description:           req.Description,
		Category:              req.Category,
		ClientName:            req.ClientName,
		Address:               req.Address,
		Locality:              req.Locality,
		City:                  req.City,
		Pincode:               req.Pincode,
		Latitude:              req.Latitude,
		Longitude:             req.Longitude,
		EstimatedDurationMins: req.EstimatedDurationMins,
		PaymentAmountRupees:   req.PaymentAmountRupees,
		ScheduledDate:         scheduledDate,
		ScheduledTimeSlot:     req.ScheduledTimeSlot,
		IsUrgent:              req.IsUrgent,
		RequiredSkills:        req.RequiredSkills,
	}

	result, err := h.jobService.CreateJob(c.Request().Context(), userID, job)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create job"})
	}

	return c.JSON(http.StatusCreated, result)
}

// GetAvailableJobs handles GET /api/v1/jobs
func (h *JobHandler) GetAvailableJobs(c echo.Context) error {
	locality := c.QueryParam("locality")
	categoryStr := c.QueryParam("category")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	var category *domain.ServiceCategory
	if categoryStr != "" {
		cat := domain.ServiceCategory(categoryStr)
		category = &cat
	}

	jobs, err := h.jobService.GetAvailableJobs(c.Request().Context(), locality, category, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get jobs"})
	}

	return c.JSON(http.StatusOK, jobs)
}

// GetMyJobs handles GET /api/v1/jobs/my
func (h *JobHandler) GetMyJobs(c echo.Context) error {
	userID := c.Get("user_id").(string)
	userType := c.Get("user_type").(string)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	jobs, err := h.jobService.GetMyJobs(c.Request().Context(), userID, userType, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get jobs"})
	}

	return c.JSON(http.StatusOK, jobs)
}

// GetJob handles GET /api/v1/jobs/:id
func (h *JobHandler) GetJob(c echo.Context) error {
	jobID := c.Param("id")

	job, err := h.jobService.GetJob(c.Request().Context(), jobID)
	if err != nil {
		if err == service.ErrJobNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get job"})
	}

	return c.JSON(http.StatusOK, job)
}

// AcceptJob handles POST /api/v1/jobs/:id/accept
func (h *JobHandler) AcceptJob(c echo.Context) error {
	jobID := c.Param("id")
	workerID := c.Get("user_id").(string)

	job, err := h.jobService.AcceptJob(c.Request().Context(), jobID, workerID)
	if err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrJobNotOpen:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job is not open"})
		case service.ErrCannotAcceptOwnJob:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot accept your own job"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to accept job"})
		}
	}

	return c.JSON(http.StatusOK, job)
}

// StartJob handles POST /api/v1/jobs/:id/start
func (h *JobHandler) StartJob(c echo.Context) error {
	jobID := c.Param("id")
	workerID := c.Get("user_id").(string)

	job, err := h.jobService.StartJob(c.Request().Context(), jobID, workerID)
	if err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrJobNotAssigned:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job is not assigned"})
		case service.ErrNotAssignedWorker:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not the assigned worker"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start job"})
		}
	}

	return c.JSON(http.StatusOK, job)
}

// CompleteJob handles POST /api/v1/jobs/:id/complete
func (h *JobHandler) CompleteJob(c echo.Context) error {
	jobID := c.Param("id")
	workerID := c.Get("user_id").(string)

	var req CompleteJobRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	job, err := h.jobService.CompleteJob(c.Request().Context(), jobID, workerID, req.Notes, req.PhotoURLs)
	if err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrInvalidTransition:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job is not in progress"})
		case service.ErrNotAssignedWorker:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not the assigned worker"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to complete job"})
		}
	}

	return c.JSON(http.StatusOK, job)
}

// ConfirmCompletion handles POST /api/v1/jobs/:id/confirm
func (h *JobHandler) ConfirmCompletion(c echo.Context) error {
	jobID := c.Param("id")
	userID := c.Get("user_id").(string)

	job, err := h.jobService.ConfirmCompletion(c.Request().Context(), jobID, userID)
	if err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrInvalidTransition:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job is not awaiting confirmation"})
		case service.ErrNotJobOwner:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only the job owner can confirm completion"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to confirm completion"})
		}
	}

	return c.JSON(http.StatusOK, job)
}

// CancelJob handles POST /api/v1/jobs/:id/cancel
func (h *JobHandler) CancelJob(c echo.Context) error {
	jobID := c.Param("id")
	userID := c.Get("user_id").(string)

	job, err := h.jobService.CancelJob(c.Request().Context(), jobID, userID)
	if err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrInvalidTransition:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job cannot be cancelled"})
		case service.ErrNotJobOwner:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not authorized to cancel this job"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to cancel job"})
		}
	}

	return c.JSON(http.StatusOK, job)
}
