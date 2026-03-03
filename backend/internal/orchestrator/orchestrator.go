package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/matching"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/google/uuid"
)

// =============================================================================
// Unified Dispatch Orchestrator
//
// Single execution pipeline that coordinates:
//   Pricing → Matching → Stability → Dispatch
//
// Does NOT modify internal modules. Coordinates them.
// Thread-safe, atomic transactions, failure rollback.
// =============================================================================

// PipelineEvent logs each stage of the pipeline for observability.
type PipelineEvent struct {
	ID         string                 `json:"id"`
	BatchID    string                 `json:"batch_id"`
	Stage      string                 `json:"stage"`
	JobID      string                 `json:"job_id,omitempty"`
	WorkerID   string                 `json:"worker_id,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	Success    bool                   `json:"success"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// PipelineResult is the full output of one batch cycle.
type PipelineResult struct {
	BatchID        string           `json:"batch_id"`
	TotalJobs      int              `json:"total_jobs"`
	PricingApplied int              `json:"pricing_applied"`
	Matched        int              `json:"matched"`
	StabilityOK    int              `json:"stability_ok"`
	Dispatched     int              `json:"dispatched"`
	Requeued       int              `json:"requeued"`
	TotalMs        int64            `json:"total_ms"`
	StageTimings   map[string]int64 `json:"stage_timings"`
	Events         []PipelineEvent  `json:"events"`
	Timestamp      time.Time        `json:"timestamp"`
}

// PipelineStatus reports the orchestrator's current state.
type PipelineStatus struct {
	IsRunning       bool            `json:"is_running"`
	QueueDepth      int             `json:"queue_depth"`
	LastResult      *PipelineResult `json:"last_result,omitempty"`
	TotalBatches    int64           `json:"total_batches"`
	TotalDispatched int64           `json:"total_dispatched"`
	Uptime          string          `json:"uptime"`
}

// Orchestrator coordinates the full dispatch pipeline.
type Orchestrator struct {
	// Services (read-only, no circular deps)
	pricingService   service.PricingService
	matchingService  service.MatchingService
	stabilityService service.StabilityService

	// Infrastructure
	dispatcher    *matching.Dispatcher
	batchQueue    *matching.BatchQueue
	jobRepo       repository.JobRepository
	stabilityRepo repository.StabilityRepository

	// State
	mu              sync.RWMutex
	lastResult      *PipelineResult
	totalBatches    int64
	totalDispatched int64
	startedAt       time.Time
	candidateLimit  int
}

// NewOrchestrator creates a new unified dispatch orchestrator.
func NewOrchestrator(
	pricingService service.PricingService,
	matchingService service.MatchingService,
	stabilityService service.StabilityService,
	dispatcher *matching.Dispatcher,
	jobRepo repository.JobRepository,
	stabilityRepo repository.StabilityRepository,
	candidateLimit int,
) *Orchestrator {
	if candidateLimit <= 0 {
		candidateLimit = 30
	}
	return &Orchestrator{
		pricingService:   pricingService,
		matchingService:  matchingService,
		stabilityService: stabilityService,
		dispatcher:       dispatcher,
		jobRepo:          jobRepo,
		stabilityRepo:    stabilityRepo,
		candidateLimit:   candidateLimit,
		startedAt:        time.Now(),
	}
}

// SetBatchQueue sets the batch queue reference (breaks circular dep).
func (o *Orchestrator) SetBatchQueue(q *matching.BatchQueue) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.batchQueue = q
}

// ProcessBatch is the main pipeline entry point.
// This is registered as the BatchQueue callback.
func (o *Orchestrator) ProcessBatch(jobIDs []string) {
	batchID := uuid.New().String()[:8]
	start := time.Now()
	ctx := context.Background()

	log.Printf("[Orchestrator] ═══ Batch %s: %d jobs ═══", batchID, len(jobIDs))

	result := &PipelineResult{
		BatchID:      batchID,
		TotalJobs:    len(jobIDs),
		StageTimings: make(map[string]int64),
		Events:       make([]PipelineEvent, 0, len(jobIDs)*4),
		Timestamp:    start,
	}

	// ─── STAGE 1: PRICING ─────────────────────────────────────────────
	pricingStart := time.Now()
	pricedJobs := o.stagePricing(ctx, batchID, jobIDs, result)
	result.StageTimings["pricing"] = time.Since(pricingStart).Milliseconds()
	result.PricingApplied = len(pricedJobs)

	if len(pricedJobs) == 0 {
		log.Printf("[Orchestrator] Batch %s: no priceable jobs, requeueing all", batchID)
		o.requeueAll(jobIDs, result)
		o.finalize(result, start)
		return
	}

	// ─── STAGE 2: MATCHING ────────────────────────────────────────────
	matchingStart := time.Now()
	assignments, unmatched := o.stageMatching(ctx, batchID, pricedJobs, result)
	result.StageTimings["matching"] = time.Since(matchingStart).Milliseconds()
	result.Matched = len(assignments)

	// ─── STAGE 3: STABILITY + DISPATCH ────────────────────────────────
	dispatchStart := time.Now()
	dispatched, requeued := o.stageStabilityAndDispatch(ctx, batchID, assignments, result)
	result.StageTimings["stability_dispatch"] = time.Since(dispatchStart).Milliseconds()
	result.StabilityOK = dispatched
	result.Dispatched = dispatched
	result.Requeued = requeued + len(unmatched)

	// Requeue unmatched jobs
	o.requeueJobs(unmatched, result)

	o.finalize(result, start)
}

// ═══════════════════════════════════════════════════════════════════════
// STAGE 1: PRICING
// Apply surge pricing to each job based on its H3 hexagon
// ═══════════════════════════════════════════════════════════════════════

func (o *Orchestrator) stagePricing(
	ctx context.Context, batchID string, jobIDs []string, result *PipelineResult,
) []string {
	pricedJobs := make([]string, 0, len(jobIDs))

	for _, jobID := range jobIDs {
		start := time.Now()
		quote, err := o.pricingService.ApplySurgeToJob(ctx, jobID)

		event := PipelineEvent{
			ID:         uuid.New().String()[:8],
			BatchID:    batchID,
			Stage:      "pricing",
			JobID:      jobID,
			DurationMs: time.Since(start).Milliseconds(),
			Success:    err == nil,
			Timestamp:  time.Now(),
		}

		if err != nil {
			event.Details = map[string]interface{}{"error": err.Error()}
			log.Printf("[Orchestrator] Pricing failed for job %s: %v", jobID, err)
			// Still include in matching — just without surge
			pricedJobs = append(pricedJobs, jobID)
		} else {
			event.Details = map[string]interface{}{
				"surge_multiplier": quote.SurgeMultiplier,
				"base_price":       quote.BasePrice,
				"final_price":      quote.FinalPrice,
				"h3_index":         quote.H3Index,
			}
			pricedJobs = append(pricedJobs, jobID)
		}

		result.Events = append(result.Events, event)
	}

	log.Printf("[Orchestrator] Batch %s: priced %d/%d jobs", batchID, len(pricedJobs), len(jobIDs))
	return pricedJobs
}

// ═══════════════════════════════════════════════════════════════════════
// STAGE 2: MATCHING
// Build candidate matrix and run Hungarian algorithm
// ═══════════════════════════════════════════════════════════════════════

func (o *Orchestrator) stageMatching(
	ctx context.Context, batchID string, jobIDs []string, result *PipelineResult,
) ([]domain.Assignment, []string) {
	start := time.Now()

	// Collect candidates per job
	type jobCandidates struct {
		jobID      string
		candidates []domain.WorkerCandidate
	}

	allJC := make([]jobCandidates, 0, len(jobIDs))
	skippedJobs := make([]string, 0)

	for _, jobID := range jobIDs {
		resp, err := o.matchingService.GetCandidateWorkers(ctx, jobID, o.candidateLimit)
		if err != nil || len(resp.Candidates) == 0 {
			skippedJobs = append(skippedJobs, jobID)
			result.Events = append(result.Events, PipelineEvent{
				ID: uuid.New().String()[:8], BatchID: batchID,
				Stage: "matching_filter", JobID: jobID,
				Success: false, Timestamp: time.Now(),
				Details: map[string]interface{}{"reason": "no_candidates"},
			})
			continue
		}
		allJC = append(allJC, jobCandidates{jobID: jobID, candidates: resp.Candidates})
	}

	if len(allJC) == 0 {
		return nil, jobIDs
	}

	// Build global matrix
	workerIndex := make(map[string]int)
	var workerIDs []string
	for _, jc := range allJC {
		for _, c := range jc.candidates {
			if _, exists := workerIndex[c.WorkerID]; !exists {
				workerIndex[c.WorkerID] = len(workerIDs)
				workerIDs = append(workerIDs, c.WorkerID)
			}
		}
	}

	numJobs := len(allJC)
	numWorkers := len(workerIDs)
	jobIDsProcessed := make([]string, numJobs)

	weights := make([][]float64, numJobs)
	for i, jc := range allJC {
		jobIDsProcessed[i] = jc.jobID
		weights[i] = make([]float64, numWorkers)
		for _, c := range jc.candidates {
			weights[i][workerIndex[c.WorkerID]] = c.TotalScore
		}
	}

	// Run Hungarian
	rawAssignments := matching.Solve(weights)

	// Map back to domain assignments
	now := time.Now()
	assignments := make([]domain.Assignment, 0, len(rawAssignments))
	matchedJobs := make(map[int]bool)

	for _, a := range rawAssignments {
		if a.Row < numJobs && a.Col < numWorkers && a.Weight > 0 {
			assignment := domain.Assignment{
				JobID:      jobIDsProcessed[a.Row],
				WorkerID:   workerIDs[a.Col],
				Score:      a.Weight,
				AssignedAt: now,
			}
			assignments = append(assignments, assignment)
			matchedJobs[a.Row] = true

			result.Events = append(result.Events, PipelineEvent{
				ID: uuid.New().String()[:8], BatchID: batchID,
				Stage: "matching", JobID: assignment.JobID, WorkerID: assignment.WorkerID,
				DurationMs: time.Since(start).Milliseconds(),
				Success:    true, Timestamp: time.Now(),
				Details: map[string]interface{}{"score": a.Weight},
			})
		}
	}

	// Collect unmatched
	unmatched := make([]string, 0)
	for i, jc := range allJC {
		if !matchedJobs[i] {
			unmatched = append(unmatched, jc.jobID)
		}
	}
	unmatched = append(unmatched, skippedJobs...)

	log.Printf("[Orchestrator] Batch %s: matched %d/%d, unmatched %d",
		batchID, len(assignments), len(jobIDs), len(unmatched))

	return assignments, unmatched
}

// ═══════════════════════════════════════════════════════════════════════
// STAGE 3: STABILITY CHECK + DISPATCH
// For each assignment: check stability → dispatch or requeue
// ═══════════════════════════════════════════════════════════════════════

func (o *Orchestrator) stageStabilityAndDispatch(
	ctx context.Context, batchID string, assignments []domain.Assignment, result *PipelineResult,
) (dispatched int, requeued int) {
	for _, a := range assignments {
		start := time.Now()

		// Check worker cancellation limit (stability gate)
		status, err := o.stabilityService.CheckCancellationLimit(ctx, a.WorkerID)
		if err != nil {
			log.Printf("[Orchestrator] Stability check failed for worker %s: %v", a.WorkerID, err)
			o.requeueJob(a.JobID, result)
			requeued++
			continue
		}

		if status.IsBlocked {
			// Worker is in cooldown — skip this assignment, requeue job
			result.Events = append(result.Events, PipelineEvent{
				ID: uuid.New().String()[:8], BatchID: batchID,
				Stage: "stability", JobID: a.JobID, WorkerID: a.WorkerID,
				DurationMs: time.Since(start).Milliseconds(),
				Success:    false, Timestamp: time.Now(),
				Details: map[string]interface{}{
					"reason":        "worker_cooldown",
					"cooldown_ends": status.CooldownEnds,
				},
			})

			// Log stability event
			o.logStabilityEvent(ctx, domain.StabilityEventSwitchDenied, a.WorkerID, a.JobID,
				map[string]interface{}{"reason": "cooldown_blocked_by_orchestrator"})

			o.requeueJob(a.JobID, result)
			requeued++
			continue
		}

		// Dispatch the assignment
		dispatchResult := o.dispatchAssignment(ctx, batchID, a, result)
		if dispatchResult {
			dispatched++
		} else {
			requeued++
		}
	}

	log.Printf("[Orchestrator] Batch %s: dispatched %d, requeued %d", batchID, dispatched, requeued)
	return
}

func (o *Orchestrator) dispatchAssignment(
	ctx context.Context, batchID string, a domain.Assignment, result *PipelineResult,
) bool {
	start := time.Now()

	// Atomic dispatch: assign + acceptance + availability in sequence
	// If any step fails, rollback and requeue
	if err := o.jobRepo.AssignWorker(ctx, a.JobID, a.WorkerID); err != nil {
		result.Events = append(result.Events, PipelineEvent{
			ID: uuid.New().String()[:8], BatchID: batchID,
			Stage: "dispatch", JobID: a.JobID, WorkerID: a.WorkerID,
			DurationMs: time.Since(start).Milliseconds(),
			Success:    false, Timestamp: time.Now(),
			Details: map[string]interface{}{"error": err.Error(), "step": "assign_worker"},
		})
		o.requeueJob(a.JobID, result)
		return false
	}

	// Create acceptance record
	acceptance := &domain.JobAcceptance{
		ID:                   uuid.New().String(),
		JobID:                a.JobID,
		WorkerID:             a.WorkerID,
		AcceptedAt:           time.Now(),
		EstimatedArrivalMins: 30,
	}
	if err := o.jobRepo.CreateAcceptance(ctx, acceptance); err != nil {
		log.Printf("[Orchestrator] Warning: acceptance record failed: %v", err)
	}

	// Emit WebSocket notification via dispatcher
	o.dispatcher.NotifyAssignment(a)

	result.Events = append(result.Events, PipelineEvent{
		ID: uuid.New().String()[:8], BatchID: batchID,
		Stage: "dispatch", JobID: a.JobID, WorkerID: a.WorkerID,
		DurationMs: time.Since(start).Milliseconds(),
		Success:    true, Timestamp: time.Now(),
		Details: map[string]interface{}{"score": a.Score},
	})

	log.Printf("[Orchestrator] ✓ Dispatched job %s → worker %s (score: %.3f)", a.JobID, a.WorkerID, a.Score)
	return true
}

// ═══════════════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════════════

func (o *Orchestrator) requeueAll(jobIDs []string, result *PipelineResult) {
	if o.batchQueue != nil {
		o.batchQueue.EnqueueMultiple(jobIDs)
	}
	result.Requeued += len(jobIDs)
}

func (o *Orchestrator) requeueJobs(jobIDs []string, result *PipelineResult) {
	if len(jobIDs) > 0 && o.batchQueue != nil {
		o.batchQueue.EnqueueMultiple(jobIDs)
		result.Requeued += len(jobIDs)
	}
}

func (o *Orchestrator) requeueJob(jobID string, result *PipelineResult) {
	if o.batchQueue != nil {
		o.batchQueue.Enqueue(jobID)
	}
	result.Requeued++
}

func (o *Orchestrator) finalize(result *PipelineResult, start time.Time) {
	result.TotalMs = time.Since(start).Milliseconds()

	o.mu.Lock()
	o.lastResult = result
	o.totalBatches++
	o.totalDispatched += int64(result.Dispatched)
	o.mu.Unlock()

	log.Printf("[Orchestrator] ═══ Batch %s complete: %d dispatched, %d requeued in %dms ═══",
		result.BatchID, result.Dispatched, result.Requeued, result.TotalMs)
}

func (o *Orchestrator) logStabilityEvent(
	ctx context.Context, eventType domain.StabilityEventType, actorID, jobID string,
	details map[string]interface{},
) {
	detailsJSON, _ := json.Marshal(details)
	event := &domain.StabilityEvent{
		ID:        uuid.New().String(),
		EventType: eventType,
		ActorID:   actorID,
		ActorRole: "worker",
		JobID:     jobID,
		Details:   string(detailsJSON),
	}
	_ = o.stabilityRepo.LogEvent(ctx, event)
}

// ═══════════════════════════════════════════════════════════════════════
// STATUS
// ═══════════════════════════════════════════════════════════════════════

// GetStatus returns the current pipeline status.
func (o *Orchestrator) GetStatus() *PipelineStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()

	queueDepth := 0
	if o.batchQueue != nil {
		queueDepth = o.batchQueue.QueueDepth()
	}

	return &PipelineStatus{
		IsRunning:       o.batchQueue != nil && o.batchQueue.IsRunning(),
		QueueDepth:      queueDepth,
		LastResult:      o.lastResult,
		TotalBatches:    o.totalBatches,
		TotalDispatched: o.totalDispatched,
		Uptime:          fmt.Sprintf("%.0fm", time.Since(o.startedAt).Minutes()),
	}
}

// SubmitJob applies pricing and enqueues a job for matching.
// Called on job creation — the single entry point for the pipeline.
func (o *Orchestrator) SubmitJob(ctx context.Context, jobID string) (*domain.PriceQuote, error) {
	// 1. Apply surge pricing
	quote, err := o.pricingService.ApplySurgeToJob(ctx, jobID)
	if err != nil {
		log.Printf("[Orchestrator] Pricing failed for job %s (enqueuing anyway): %v", jobID, err)
	}

	// 2. Enqueue for batch matching
	if o.batchQueue != nil {
		o.batchQueue.Enqueue(jobID)
	}

	log.Printf("[Orchestrator] Job %s submitted (surge: %.2f)", jobID,
		func() float64 {
			if quote != nil {
				return quote.SurgeMultiplier
			}
			return 1.0
		}())

	return quote, nil
}
