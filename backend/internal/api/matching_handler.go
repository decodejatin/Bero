package api

import (
	"net/http"
	"strconv"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
)

// MatchingHandler handles the matching pre-processing API.
type MatchingHandler struct {
	matchingService service.MatchingService
}

// NewMatchingHandler creates a new matching handler.
func NewMatchingHandler(matchingService service.MatchingService) *MatchingHandler {
	return &MatchingHandler{matchingService: matchingService}
}

// --- Request / Response types ---

// UpdateWeightsRequest is the body for PUT /matching/weights
type UpdateWeightsRequest struct {
	DistanceWeight    float64 `json:"distance_weight"`
	ReputationWeight  float64 `json:"reputation_weight"`
	SkillWeight       float64 `json:"skill_weight"`
	WaitPenaltyWeight float64 `json:"wait_penalty_weight"`
}

// BuildMatrixRequest is the body for POST /matching/matrix
type BuildMatrixRequest struct {
	JobIDs         []string `json:"job_ids"`
	CandidateLimit int      `json:"candidate_limit"`
}

// --- Handlers ---

// GetCandidates handles GET /api/v1/matching/candidates/:jobId
// Returns scored candidate workers for a specific job.
func (h *MatchingHandler) GetCandidates(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "jobId is required"})
	}

	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	result, err := h.matchingService.GetCandidateWorkers(c.Request().Context(), jobID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GetWeights handles GET /api/v1/matching/weights
// Returns the current dynamic matching weight configuration.
func (h *MatchingHandler) GetWeights(c echo.Context) error {
	weights, err := h.matchingService.GetWeights(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load weights"})
	}

	return c.JSON(http.StatusOK, weights)
}

// UpdateWeights handles PUT /api/v1/matching/weights
// Updates matching weights at runtime — no redeploy needed.
func (h *MatchingHandler) UpdateWeights(c echo.Context) error {
	var req UpdateWeightsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate weights are positive
	if req.DistanceWeight < 0 || req.ReputationWeight < 0 || req.SkillWeight < 0 || req.WaitPenaltyWeight < 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "all weights must be non-negative"})
	}

	// Validate weights sum to ~1.0 (allow small tolerance)
	total := req.DistanceWeight + req.ReputationWeight + req.SkillWeight + req.WaitPenaltyWeight
	if total < 0.01 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "weights must sum to a positive value"})
	}

	weights := &domain.MatchingWeights{
		DistanceWeight:    req.DistanceWeight,
		ReputationWeight:  req.ReputationWeight,
		SkillWeight:       req.SkillWeight,
		WaitPenaltyWeight: req.WaitPenaltyWeight,
	}

	if err := h.matchingService.UpdateWeights(c.Request().Context(), weights); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update weights"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"weights": weights,
	})
}

// BuildMatrix handles POST /api/v1/matching/matrix
// Builds the weight matrix for a batch of jobs.
// This is the input for the Hungarian algorithm.
func (h *MatchingHandler) BuildMatrix(c echo.Context) error {
	var req BuildMatrixRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if len(req.JobIDs) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job_ids array is required"})
	}

	if len(req.JobIDs) > 50 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "maximum 50 jobs per batch"})
	}

	limit := req.CandidateLimit
	if limit <= 0 {
		limit = 20
	}

	matrix, err := h.matchingService.BuildWeightMatrix(c.Request().Context(), req.JobIDs, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, matrix)
}
