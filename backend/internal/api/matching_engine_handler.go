package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/matching"
	"github.com/labstack/echo/v4"
)

// MatchingEngineHandler handles the batched Hungarian matching engine API.
type MatchingEngineHandler struct {
	batchQueue *matching.BatchQueue
	dispatcher *matching.Dispatcher
}

// NewMatchingEngineHandler creates a new matching engine handler.
func NewMatchingEngineHandler(
	batchQueue *matching.BatchQueue,
	dispatcher *matching.Dispatcher,
) *MatchingEngineHandler {
	return &MatchingEngineHandler{
		batchQueue: batchQueue,
		dispatcher: dispatcher,
	}
}

// --- Request types ---

// EnqueueRequest is the body for POST /matching/enqueue.
type EnqueueRequest struct {
	JobID string `json:"job_id"`
}

// --- Handlers ---

// EnqueueJob handles POST /api/v1/matching/enqueue
// Adds a job to the matching queue for the next batch cycle.
func (h *MatchingEngineHandler) EnqueueJob(c echo.Context) error {
	var req EnqueueRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.JobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job_id is required"})
	}

	h.batchQueue.Enqueue(req.JobID)

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"status":      "queued",
		"job_id":      req.JobID,
		"queue_depth": h.batchQueue.QueueDepth(),
	})
}

// GetQueueStatus handles GET /api/v1/matching/queue/status
// Returns current queue depth and last batch result.
func (h *MatchingEngineHandler) GetQueueStatus(c echo.Context) error {
	status := domain.BatchStatus{
		QueueDepth:     h.batchQueue.QueueDepth(),
		LastBatchAt:    h.batchQueue.LastBatchAt(),
		LastResult:     h.dispatcher.LastResult(),
		IsRunning:      h.batchQueue.IsRunning(),
		BatchIntervalS: 20,
	}

	return c.JSON(http.StatusOK, status)
}

// TriggerBatch handles POST /api/v1/matching/batch/trigger
// Force-triggers an immediate batch processing cycle (admin use).
func (h *MatchingEngineHandler) TriggerBatch(c echo.Context) error {
	if h.batchQueue.QueueDepth() == 0 {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "empty_queue",
		})
	}

	h.batchQueue.TriggerNow()

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"status":      "triggered",
		"queue_depth": h.batchQueue.QueueDepth(),
	})
}

// DeclineJob handles POST /api/v1/matching/decline/:jobId
// Worker declines an assignment → job requeued for next batch.
func (h *MatchingEngineHandler) DeclineJob(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "jobId is required"})
	}

	if err := h.dispatcher.DeclineJob(c.Request().Context(), jobID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "declined_and_requeued",
		"job_id": jobID,
	})
}

// =============================================================================
// WebSocket Notification Hub Extension
// Adds SendToUser capability to the existing WebSocketHub pattern.
// =============================================================================

// NotificationHub wraps the ability to send targeted messages to specific users.
// This is injected into the Dispatcher via the NotifyFunc callback.
type NotificationHub struct {
	mu      sync.RWMutex
	clients map[string]chan []byte // userID → send channel
}

// NewNotificationHub creates a notification hub.
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[string]chan []byte),
	}
}

// RegisterClient adds a client's send channel.
func (nh *NotificationHub) RegisterClient(userID string, sendCh chan []byte) {
	nh.mu.Lock()
	defer nh.mu.Unlock()
	nh.clients[userID] = sendCh
}

// UnregisterClient removes a client.
func (nh *NotificationHub) UnregisterClient(userID string) {
	nh.mu.Lock()
	defer nh.mu.Unlock()
	delete(nh.clients, userID)
}

// SendToUser sends a JSON message to a specific user if connected.
func (nh *NotificationHub) SendToUser(userID string, payload []byte) {
	nh.mu.RLock()
	defer nh.mu.RUnlock()

	if ch, ok := nh.clients[userID]; ok {
		select {
		case ch <- payload:
		default:
			// Client buffer full, skip
		}
	}
}

// NotifyFunc returns a matching.NotifyFunc that sends to the hub.
func (nh *NotificationHub) NotifyFunc() matching.NotifyFunc {
	return func(userID string, payload []byte) {
		nh.SendToUser(userID, payload)
	}
}

// BroadcastMatchingEvent sends a matching event to all connected clients.
// Useful for admin dashboards monitoring batch activity.
func (nh *NotificationHub) BroadcastMatchingEvent(event interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	nh.mu.RLock()
	defer nh.mu.RUnlock()

	for _, ch := range nh.clients {
		select {
		case ch <- data:
		default:
		}
	}
}
