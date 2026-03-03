package matchmaker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/decodejatin/bero-backend/pkg/circuitbreaker"
	"github.com/decodejatin/bero-backend/pkg/distlock"
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
	Running         bool      `json:"running"`
	LastRunAt       time.Time `json:"last_run_at"`
	TotalRounds     int64     `json:"total_rounds"`
	TotalMatches    int64     `json:"total_matches"`
	WindowID        int64     `json:"window_id"`
	CircuitState    string    `json:"circuit_state"`     // CLOSED, OPEN, HALF_OPEN
	LastShadowDelta float64   `json:"last_shadow_delta"` // surplus delta from last shadow run
}

// Engine is the matching engine that runs a batching window loop.
type Engine struct {
	mu sync.RWMutex

	config MatchConfig
	status EngineStatus

	fetchWorkers WorkerFetcher
	fetchJobs    JobFetcher
	onAssign     AssignmentCallback

	// Infrastructure
	locker  distlock.DistLock              // distributed locking for assignments
	breaker *circuitbreaker.CircuitBreaker // circuit breaker for matching

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
		locker:       distlock.NewInMemoryDistLock(),
		breaker:      circuitbreaker.New(circuitbreaker.DefaultConfig()),
		triggerCh:    make(chan struct{}, 1),
		resultsCh:    make(chan MatchResult, 100),
	}
}

// SetDistLock replaces the default in-memory lock with a production implementation (e.g. Redis Redlock).
func (e *Engine) SetDistLock(dl distlock.DistLock) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.locker = dl
}

// SetCircuitBreaker replaces the circuit breaker configuration.
func (e *Engine) SetCircuitBreaker(cb *circuitbreaker.CircuitBreaker) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.breaker = cb
}

// Results returns a read-only channel of match results.
func (e *Engine) Results() <-chan MatchResult {
	return e.resultsCh
}

// Status returns the current engine status.
func (e *Engine) Status() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s := e.status
	s.CircuitState = e.breaker.State().String()
	return s
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

	var assignments []Assignment

	// 3. Run matching through circuit breaker
	cbErr := e.breaker.Execute(func() error {
		if cfg.EnablePruning && cfg.KNearestNeighbors > 0 {
			// 3a. Dynamic Candidate Pruning: use H3 k-ring spatial index
			assignments = PruneAndMatch(workers, jobs, cfg, now)
			log.Printf("[matchmaker] Window %d: %d workers × %d jobs → %d assignments (pruned, k=%d)",
				windowID, len(workers), len(jobs), len(assignments), cfg.KNearestNeighbors)
		} else {
			// 3b. Full matrix: classic Hungarian on all W×J pairs
			weightMatrix := BuildWeightMatrix(workers, jobs, cfg, now)
			assignments = SolveMaxWeightMatching(weightMatrix, workers, jobs, cfg.MinWeightThreshold)
			log.Printf("[matchmaker] Window %d: %d workers × %d jobs → %d assignments (full matrix)",
				windowID, len(workers), len(jobs), len(assignments))
		}
		return nil
	})

	if cbErr != nil {
		// Circuit breaker is OPEN — fallback to simple greedy matching
		log.Printf("[matchmaker] Window %d: circuit breaker OPEN, falling back to greedy matching", windowID)
		assignments = greedyMatch(workers, jobs, cfg, now)
		log.Printf("[matchmaker] Window %d: %d assignments (greedy fallback)", windowID, len(assignments))
	}

	// 3.5 Enforce stability (§3.1.2 — Modified Gale-Shapley)
	if cfg.EnableStability {
		assignments = EnforceStability(assignments, workers, jobs, cfg, now)
		log.Printf("[matchmaker] Window %d: %d stable assignments after Gale-Shapley refinement",
			windowID, len(assignments))
	}

	// 3.6 Shadow matching (A/B testing — never affects live assignments)
	if cfg.EnableShadowMode {
		shadowResult := RunShadowMatching(workers, jobs, cfg, now, assignments)
		e.mu.Lock()
		e.status.LastShadowDelta = shadowResult.Comparison.WeightDeltaPct
		e.mu.Unlock()
	}

	// 4. Execute assignments via callback with distributed locking
	for _, a := range assignments {
		// Lock the worker to prevent double-booking across concurrent processes
		lock, lockErr := e.locker.Acquire(ctx, fmt.Sprintf("assign:worker:%s", a.WorkerID), 10*time.Second)
		if lockErr != nil {
			log.Printf("[matchmaker] Cannot lock worker %s — skipping (already locked): %v", a.WorkerID, lockErr)
			continue
		}

		if err := e.onAssign(ctx, a); err != nil {
			log.Printf("[matchmaker] Failed to assign worker %s to job %s: %v", a.WorkerID, a.JobID, err)
		}

		lock.Release(ctx)
	}

	// 5. Publish result
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

	// 6. Update status
	e.mu.Lock()
	e.status.TotalRounds++
	e.status.TotalMatches += int64(len(assignments))
	e.status.LastRunAt = now
	e.mu.Unlock()
}
