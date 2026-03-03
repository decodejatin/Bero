package matching

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/decodejatin/bero-backend/internal/domain"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/google/uuid"
)

// =============================================================================
// Dispatcher — Handles post-matching assignment and failure recovery.
//
// Responsibilities:
//   - Assign worker to job via existing repo methods
//   - Mark worker as unavailable
//   - Send WebSocket notification
//   - Start decline timeout
//   - Requeue on failure
// =============================================================================

// NotifyFunc sends a real-time notification to a user.
// Injected by the API layer (wraps WebSocketHub.SendToUser).
type NotifyFunc func(userID string, payload []byte)

// Dispatcher handles post-matching assignment and failure recovery.
type Dispatcher struct {
	jobRepo      repository.JobRepository
	locationRepo repository.LocationRepository
	notifyFunc   NotifyFunc
	batchQueue   *BatchQueue

	// Track pending assignments for timeout
	mu         sync.Mutex
	pending    map[string]*pendingAssignment // jobID → pending
	lastResult *domain.BatchResult
}

type pendingAssignment struct {
	assignment domain.Assignment
	timer      *time.Timer
}

// NewDispatcher creates a new dispatcher.
func NewDispatcher(
	jobRepo repository.JobRepository,
	locationRepo repository.LocationRepository,
	notifyFunc NotifyFunc,
	batchQueue *BatchQueue,
) *Dispatcher {
	return &Dispatcher{
		jobRepo:      jobRepo,
		locationRepo: locationRepo,
		notifyFunc:   notifyFunc,
		batchQueue:   batchQueue,
		pending:      make(map[string]*pendingAssignment),
	}
}

// SetBatchQueue sets the batch queue reference.
// Used to break circular dependency during initialization:
// Dispatcher needs BatchQueue for requeue, BatchQueue needs BatchProcessor,
// BatchProcessor needs Dispatcher for dispatch.
func (d *Dispatcher) SetBatchQueue(q *BatchQueue) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.batchQueue = q
}

// DispatchAssignments processes all assignments from a batch result.
func (d *Dispatcher) DispatchAssignments(ctx context.Context, result *domain.BatchResult) {
	d.mu.Lock()
	d.lastResult = result
	d.mu.Unlock()

	for _, a := range result.Assignments {
		if err := d.dispatchOne(ctx, a); err != nil {
			log.Printf("[Dispatcher] Failed to dispatch job %s → worker %s: %v", a.JobID, a.WorkerID, err)
			// Requeue failed assignment
			d.batchQueue.Enqueue(a.JobID)
		}
	}

	// Requeue unmatched jobs
	if len(result.UnmatchedJobs) > 0 {
		d.batchQueue.EnqueueMultiple(result.UnmatchedJobs)
		log.Printf("[Dispatcher] Requeued %d unmatched jobs", len(result.UnmatchedJobs))
	}
}

func (d *Dispatcher) dispatchOne(ctx context.Context, a domain.Assignment) error {
	// 1. Assign worker to job (sets status → ASSIGNED)
	if err := d.jobRepo.AssignWorker(ctx, a.JobID, a.WorkerID); err != nil {
		return err
	}

	// 2. Create acceptance record
	acceptance := &domain.JobAcceptance{
		ID:                   uuid.New().String(),
		JobID:                a.JobID,
		WorkerID:             a.WorkerID,
		AcceptedAt:           time.Now(),
		EstimatedArrivalMins: 30,
	}
	if err := d.jobRepo.CreateAcceptance(ctx, acceptance); err != nil {
		log.Printf("[Dispatcher] Warning: acceptance record failed for job %s: %v", a.JobID, err)
	}

	// 3. Mark worker as unavailable (remove from candidate pool)
	if err := d.locationRepo.SetWorkerAvailability(ctx, a.WorkerID, false); err != nil {
		log.Printf("[Dispatcher] Warning: availability update failed for worker %s: %v", a.WorkerID, err)
	}

	// 4. Send WebSocket notifications
	d.NotifyAssignment(a)

	// 5. Start decline timeout (60 seconds)
	d.startDeclineTimeout(a)

	log.Printf("[Dispatcher] Assigned job %s → worker %s (score: %.3f)", a.JobID, a.WorkerID, a.Score)
	return nil
}

// NotifyAssignment sends WebSocket notifications for an assignment.
// Exported for use by the Orchestrator.
func (d *Dispatcher) NotifyAssignment(a domain.Assignment) {
	if d.notifyFunc == nil {
		return
	}

	payload := map[string]interface{}{
		"type":        "job_assigned",
		"job_id":      a.JobID,
		"worker_id":   a.WorkerID,
		"score":       a.Score,
		"assigned_at": a.AssignedAt,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Notify the worker
	d.notifyFunc(a.WorkerID, data)

	// Notify the client (job owner) — get from job record
	job, err := d.jobRepo.GetByID(context.Background(), a.JobID)
	if err == nil {
		d.notifyFunc(job.ClientID, data)
	}
}

func (d *Dispatcher) startDeclineTimeout(a domain.Assignment) {
	d.mu.Lock()
	defer d.mu.Unlock()

	timer := time.AfterFunc(60*time.Second, func() {
		d.mu.Lock()
		delete(d.pending, a.JobID)
		d.mu.Unlock()

		log.Printf("[Dispatcher] Assignment timeout: job %s, worker %s", a.JobID, a.WorkerID)
		// Note: In production, check if worker actually started.
		// For now, just log. The job status is already ASSIGNED.
	})

	d.pending[a.JobID] = &pendingAssignment{
		assignment: a,
		timer:      timer,
	}
}

// DeclineJob handles a worker declining an assignment.
// Resets job status and requeues for next batch.
func (d *Dispatcher) DeclineJob(ctx context.Context, jobID string) error {
	d.mu.Lock()
	pa, ok := d.pending[jobID]
	if ok {
		pa.timer.Stop()
		delete(d.pending, jobID)
	}
	d.mu.Unlock()

	// Reset job to OPEN status
	job, err := d.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.Status = domain.JobStatusOpen
	job.AssignedWorkerID = nil
	if err := d.jobRepo.Update(ctx, job); err != nil {
		return err
	}

	// Re-enable worker availability
	if ok {
		_ = d.locationRepo.SetWorkerAvailability(ctx, pa.assignment.WorkerID, true)
	}

	// Requeue for next batch
	d.batchQueue.Enqueue(jobID)
	log.Printf("[Dispatcher] Job %s declined, requeued", jobID)

	return nil
}

// LastResult returns the most recent batch result.
func (d *Dispatcher) LastResult() *domain.BatchResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.lastResult
}

// =============================================================================
// BatchProcessor — Orchestrates the full matching pipeline.
// =============================================================================

// BatchProcessor connects candidate filtering → matrix building → Hungarian → dispatch.
type BatchProcessor struct {
	matchingService service.MatchingService
	dispatcher      *Dispatcher
	candidateLimit  int
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor(
	matchingService service.MatchingService,
	dispatcher *Dispatcher,
	candidateLimit int,
) *BatchProcessor {
	if candidateLimit <= 0 {
		candidateLimit = 30
	}
	return &BatchProcessor{
		matchingService: matchingService,
		dispatcher:      dispatcher,
		candidateLimit:  candidateLimit,
	}
}

// ProcessBatch is the main batch processing pipeline.
// Called by BatchQueue on each tick.
func (p *BatchProcessor) ProcessBatch(jobIDs []string) {
	start := time.Now()
	ctx := context.Background()

	log.Printf("[BatchProcessor] Processing %d jobs", len(jobIDs))

	// 1. Collect candidates per job using Step 2's scoring
	type jobCandidates struct {
		jobID      string
		candidates []domain.WorkerCandidate
	}

	allJC := make([]jobCandidates, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		resp, err := p.matchingService.GetCandidateWorkers(ctx, jobID, p.candidateLimit)
		if err != nil {
			log.Printf("[BatchProcessor] Skip job %s: %v", jobID, err)
			continue
		}
		if len(resp.Candidates) == 0 {
			log.Printf("[BatchProcessor] No candidates for job %s", jobID)
			continue
		}
		allJC = append(allJC, jobCandidates{
			jobID:      jobID,
			candidates: resp.Candidates,
		})
	}

	if len(allJC) == 0 {
		log.Println("[BatchProcessor] No viable jobs in batch")
		result := &domain.BatchResult{
			UnmatchedJobs:  jobIDs,
			TotalJobs:      len(jobIDs),
			TotalMatched:   0,
			ProcessingMs:   time.Since(start).Milliseconds(),
			BatchTimestamp: start,
		}
		p.dispatcher.DispatchAssignments(ctx, result)
		return
	}

	// 2. Build unique worker index + weight matrix
	//    Rows = jobs, Columns = unique workers across all jobs
	workerIndex := make(map[string]int) // workerID → column index
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

	// Build weight matrix [numJobs × numWorkers]
	weights := make([][]float64, numJobs)
	for i, jc := range allJC {
		jobIDsProcessed[i] = jc.jobID
		weights[i] = make([]float64, numWorkers)
		// Initialize with -1 (invalid pairs)
		for j := range weights[i] {
			weights[i][j] = 0 // Workers not in candidate set get zero score
		}
		// Fill in actual scores
		for _, c := range jc.candidates {
			col := workerIndex[c.WorkerID]
			weights[i][col] = c.TotalScore
		}
	}

	// 3. Run Hungarian algorithm
	assignments := Solve(weights)

	// 4. Map assignments back to job/worker IDs
	now := time.Now()
	domainAssignments := make([]domain.Assignment, 0, len(assignments))
	matchedJobs := make(map[int]bool)

	for _, a := range assignments {
		if a.Row < numJobs && a.Col < numWorkers && a.Weight > 0 {
			domainAssignments = append(domainAssignments, domain.Assignment{
				JobID:      jobIDsProcessed[a.Row],
				WorkerID:   workerIDs[a.Col],
				Score:      a.Weight,
				AssignedAt: now,
			})
			matchedJobs[a.Row] = true
		}
	}

	// 5. Identify unmatched jobs
	unmatched := make([]string, 0)
	for i, jc := range allJC {
		if !matchedJobs[i] {
			unmatched = append(unmatched, jc.jobID)
		}
	}
	// Also include jobs that were skipped (no candidates)
	processedSet := make(map[string]bool)
	for _, jc := range allJC {
		processedSet[jc.jobID] = true
	}
	for _, id := range jobIDs {
		if !processedSet[id] {
			unmatched = append(unmatched, id)
		}
	}

	result := &domain.BatchResult{
		Assignments:    domainAssignments,
		UnmatchedJobs:  unmatched,
		TotalJobs:      len(jobIDs),
		TotalMatched:   len(domainAssignments),
		ProcessingMs:   time.Since(start).Milliseconds(),
		BatchTimestamp: start,
	}

	log.Printf("[BatchProcessor] Batch complete: %d/%d matched in %dms",
		result.TotalMatched, result.TotalJobs, result.ProcessingMs)

	// 6. Dispatch assignments
	p.dispatcher.DispatchAssignments(ctx, result)
}
