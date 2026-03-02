package matchmaker

import (
	"context"
	"log"
	"sync"
	"time"
)

// WorkerFetcher is a function that retrieves online workers for matching.
// This decouples the engine from any database dependency.
type WorkerFetcher func(ctx context.Context) ([]MatchableWorker, error)

// JobFetcher is a function that retrieves unmatched open jobs for matching.
type JobFetcher func(ctx context.Context) ([]MatchableJob, error)

// AssignmentCallback is called for each successful assignment.
type AssignmentCallback func(ctx context.Context, assignment Assignment) error

// EngineStatus reports the current state of the matching engine.
type EngineStatus struct {
	Running      bool      `json:"running"`
	LastRunAt    time.Time `json:"last_run_at"`
	TotalRounds  int64     `json:"total_rounds"`
	TotalMatches int64     `json:"total_matches"`
	WindowID     int64     `json:"window_id"`
}

// Engine is the matching engine that runs a batching window loop.
type Engine struct {
	mu sync.RWMutex

	config MatchConfig
	status EngineStatus

	fetchWorkers WorkerFetcher
	fetchJobs    JobFetcher
	onAssign     AssignmentCallback

	// Manual trigger channel
	triggerCh chan struct{}

	// Results are published here for optional consumption
	resultsCh chan MatchResult
}

// NewEngine creates a new matching engine.
func NewEngine(cfg MatchConfig, fetchWorkers WorkerFetcher, fetchJobs JobFetcher, onAssign AssignmentCallback) *Engine {
	return &Engine{
		config:       cfg,
		fetchWorkers: fetchWorkers,
		fetchJobs:    fetchJobs,
		onAssign:     onAssign,
		triggerCh:    make(chan struct{}, 1),
		resultsCh:    make(chan MatchResult, 100),
	}
}

// Results returns a read-only channel of match results.
func (e *Engine) Results() <-chan MatchResult {
	return e.resultsCh
}

// Status returns the current engine status.
func (e *Engine) Status() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// UpdateConfig updates the engine configuration.
func (e *Engine) UpdateConfig(cfg MatchConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg
}

// Config returns the current engine configuration.
func (e *Engine) Config() MatchConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// Trigger manually triggers a matching round (non-blocking).
func (e *Engine) Trigger() {
	select {
	case e.triggerCh <- struct{}{}:
	default:
		// Already triggered, skip
	}
}

// Run starts the batching window loop. Blocks until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) {
	e.mu.Lock()
	e.status.Running = true
	windowDuration := time.Duration(e.config.WindowDurationSeconds) * time.Second
	e.mu.Unlock()

	log.Printf("[matchmaker] Engine started with %v window", windowDuration)

	ticker := time.NewTicker(windowDuration)
	defer ticker.Stop()
	defer func() {
		e.mu.Lock()
		e.status.Running = false
		e.mu.Unlock()
		close(e.resultsCh)
		log.Println("[matchmaker] Engine stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.runMatchingRound(ctx)
		case <-e.triggerCh:
			e.runMatchingRound(ctx)
		}
	}
}

// runMatchingRound executes a single matching round.
func (e *Engine) runMatchingRound(ctx context.Context) {
	e.mu.Lock()
	cfg := e.config
	e.status.WindowID++
	windowID := e.status.WindowID
	e.mu.Unlock()

	now := time.Now()

	// 1. Fetch available workers
	workers, err := e.fetchWorkers(ctx)
	if err != nil {
		log.Printf("[matchmaker] Error fetching workers: %v", err)
		return
	}

	// 2. Fetch unmatched open jobs
	jobs, err := e.fetchJobs(ctx)
	if err != nil {
		log.Printf("[matchmaker] Error fetching jobs: %v", err)
		return
	}

	if len(workers) == 0 || len(jobs) == 0 {
		log.Printf("[matchmaker] Window %d: %d workers, %d jobs — skipping", windowID, len(workers), len(jobs))
		e.mu.Lock()
		e.status.TotalRounds++
		e.status.LastRunAt = now
		e.mu.Unlock()
		return
	}

	// 3. Build weight matrix
	weightMatrix := BuildWeightMatrix(workers, jobs, cfg, now)

	// 4. Solve maximum weight matching via Hungarian algorithm
	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, cfg.MinWeightThreshold)

	log.Printf("[matchmaker] Window %d: %d workers × %d jobs → %d assignments", windowID, len(workers), len(jobs), len(assignments))

	// 5. Execute assignments via callback
	for _, a := range assignments {
		if err := e.onAssign(ctx, a); err != nil {
			log.Printf("[matchmaker] Failed to assign worker %s to job %s: %v", a.WorkerID, a.JobID, err)
		}
	}

	// 6. Publish result
	result := MatchResult{
		Assignments: assignments,
		Timestamp:   now,
		WindowID:    windowID,
	}

	select {
	case e.resultsCh <- result:
	default:
		// Channel full, discard oldest (consumer too slow)
	}

	// 7. Update status
	e.mu.Lock()
	e.status.TotalRounds++
	e.status.TotalMatches += int64(len(assignments))
	e.status.LastRunAt = now
	e.mu.Unlock()
}
