package api

import (
	"net/http"
	"strconv"

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

// CreateJob godoc
// @Summary Create a new job
// @Description Creates a new job listing (client only)
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body service.CreateJobRequest true "Job details"
// @Success 201 {object} domain.Job
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /jobs [post]
func (h *JobHandler) CreateJob(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req service.CreateJobRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	job, err := h.jobService.CreateJob(c.Request().Context(), userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create job"})
	}

	return c.JSON(http.StatusCreated, job)
}

// GetAvailableJobs godoc
// @Summary Get available jobs
// @Description Returns list of open jobs for workers
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param locality query string false "Filter by locality"
// @Param category query string false "Filter by category"
// @Param limit query int false "Number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} domain.Job
// @Router /jobs [get]
func (h *JobHandler) GetAvailableJobs(c echo.Context) error {
	locality := c.QueryParam("locality")
	categoryStr := c.QueryParam("category")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var category *domain.ServiceCategory
	if categoryStr != "" {
		cat := domain.ServiceCategory(categoryStr)
		category = &cat
	}

	jobs, err := h.jobService.GetAvailableJobs(c.Request().Context(), locality, category, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch jobs"})
	}

	return c.JSON(http.StatusOK, jobs)
}

// GetJob godoc
// @Summary Get job details
// @Description Returns job details by ID
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} domain.Job
// @Failure 404 {object} ErrorResponse
// @Router /jobs/{id} [get]
func (h *JobHandler) GetJob(c echo.Context) error {
	jobID := c.Param("id")

	job, err := h.jobService.GetJob(c.Request().Context(), jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
	}

	return c.JSON(http.StatusOK, job)
}

// GetMyJobs godoc
// @Summary Get my jobs
// @Description Returns jobs for the authenticated user (client or worker)
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} domain.Job
// @Router /jobs/my [get]
func (h *JobHandler) GetMyJobs(c echo.Context) error {
	userID := c.Get("user_id").(string)
	userType := c.Get("user_type").(string)
	statusStr := c.QueryParam("status")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 {
		limit = 20
	}

	var status *domain.JobStatus
	if statusStr != "" {
		s := domain.JobStatus(statusStr)
		status = &s
	}

	var jobs []domain.Job
	var err error

	if userType == string(domain.UserTypeWorker) {
		jobs, err = h.jobService.GetWorkerJobs(c.Request().Context(), userID, status, limit, offset)
	} else {
		jobs, err = h.jobService.GetClientJobs(c.Request().Context(), userID, limit, offset)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch jobs"})
	}

	return c.JSON(http.StatusOK, jobs)
}

// AcceptJobRequest is the request for accepting a job
type AcceptJobRequest struct {
	EstimatedArrivalMins int `json:"estimated_arrival_minutes"`
}

// AcceptJob godoc
// @Summary Accept a job
// @Description Worker accepts an open job
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body AcceptJobRequest true "Acceptance details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /jobs/{id}/accept [post]
func (h *JobHandler) AcceptJob(c echo.Context) error {
	userID := c.Get("user_id").(string)
	jobID := c.Param("id")

	var req AcceptJobRequest
	if err := c.Bind(&req); err != nil {
		req.EstimatedArrivalMins = 30
	}

	if err := h.jobService.AcceptJob(c.Request().Context(), userID, jobID, req.EstimatedArrivalMins); err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrAlreadyAssigned:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job already assigned"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to accept job"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "job accepted successfully"})
}

// StartJob godoc
// @Summary Start a job
// @Description Worker starts working on an assigned job
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /jobs/{id}/start [post]
func (h *JobHandler) StartJob(c echo.Context) error {
	userID := c.Get("user_id").(string)
	jobID := c.Param("id")

	if err := h.jobService.StartJob(c.Request().Context(), userID, jobID); err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrUnauthorizedAction:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		case service.ErrInvalidStatus:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status transition"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to start job"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "job started"})
}

// CompleteJob godoc
// @Summary Complete a job
// @Description Worker marks job as completed
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body service.CompleteJobRequest true "Completion details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /jobs/{id}/complete [post]
func (h *JobHandler) CompleteJob(c echo.Context) error {
	userID := c.Get("user_id").(string)
	jobID := c.Param("id")

	var req service.CompleteJobRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if err := h.jobService.CompleteJob(c.Request().Context(), userID, jobID, &req); err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrUnauthorizedAction:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		case service.ErrInvalidStatus:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status transition"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to complete job"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "job completed"})
}

// CancelJob godoc
// @Summary Cancel a job
// @Description Client cancels a job
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /jobs/{id}/cancel [post]
func (h *JobHandler) CancelJob(c echo.Context) error {
	userID := c.Get("user_id").(string)
	jobID := c.Param("id")

	if err := h.jobService.CancelJob(c.Request().Context(), userID, jobID); err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrUnauthorizedAction:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		case service.ErrInvalidStatus:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot cancel job in current status"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to cancel job"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "job cancelled"})
}

// ConfirmCompletion godoc
// @Summary Confirm job completion
// @Description Client confirms that job is completed
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /jobs/{id}/confirm [post]
func (h *JobHandler) ConfirmCompletion(c echo.Context) error {
	userID := c.Get("user_id").(string)
	jobID := c.Param("id")

	if err := h.jobService.ConfirmCompletion(c.Request().Context(), userID, jobID); err != nil {
		switch err {
		case service.ErrJobNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "job not found"})
		case service.ErrUnauthorizedAction:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		case service.ErrInvalidStatus:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job not awaiting confirmation"})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to confirm completion"})
		}
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "job completion confirmed"})
}
