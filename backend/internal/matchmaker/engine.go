package matchmaker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/decodejatin/bero-backend/pkg/circuitbreaker"
	"github.com/decodejatin/bero-backend/pkg/distlock"
	"github.com/decodejatin/bero-backend/pkg/ratelimiter"
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
	PressureTier    string    `json:"pressure_tier"`     // NORMAL, HIGH, CRITICAL
	RoundLatencyMs  int64     `json:"round_latency_ms"`  // last round duration in milliseconds
	DroppedRounds   int64     `json:"dropped_rounds"`    // rounds dropped by rate limiter
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
	locker   distlock.DistLock              // distributed locking for assignments
	breaker  *circuitbreaker.CircuitBreaker // circuit breaker for matching
	limiter  ratelimiter.Limiter            // token bucket admission gate
	pressure *PressureMonitor               // load-aware tier tracker

	// Manual trigger channel
	triggerCh chan struct{}

	// Results are published here for optional consumption
	resultsCh chan MatchResult
}

// NewEngine creates a new matching engine with all safety systems.
func NewEngine(cfg MatchConfig, fetchWorkers WorkerFetcher, fetchJobs JobFetcher, onAssign AssignmentCallback) *Engine {
	// Initialize rate limiter
	var lim ratelimiter.Limiter
	if cfg.EnableRateLimiting {
		lim = ratelimiter.NewInMemoryLimiter(cfg.RateLimitPerSecond, cfg.RateLimitBurst)
	} else {
		lim = ratelimiter.NewNoopLimiter()
	}

	// Initialize pressure monitor from config
	pressureCfg := PressureConfig{
		HighQueueDepth:     cfg.HighPressureQueueDepth,
		CriticalQueueDepth: cfg.CritPressureQueueDepth,
		HighLatency:        time.Duration(cfg.HighPressureLatencyMs) * time.Millisecond,
		CriticalLatency:    time.Duration(cfg.CritPressureLatencyMs) * time.Millisecond,
	}
	if pressureCfg.HighQueueDepth == 0 {
		pressureCfg = DefaultPressureConfig()
	}

	return &Engine{
		config:       cfg,
		fetchWorkers: fetchWorkers,
		fetchJobs:    fetchJobs,
		onAssign:     onAssign,
		locker:       distlock.NewInMemoryDistLock(),
		breaker:      circuitbreaker.New(circuitbreaker.DefaultConfig()),
		limiter:      lim,
		pressure:     NewPressureMonitor(pressureCfg),
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

// SetRateLimiter replaces the rate limiter (e.g., swap to Redis-based in production).
func (e *Engine) SetRateLimiter(lim ratelimiter.Limiter) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.limiter = lim
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
	s.PressureTier = e.pressure.Tier().String()
	return s
}

// UpdateConfig updates the engine configuration (including rate limiter hot-reload).
func (e *Engine) UpdateConfig(cfg MatchConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg

	// Hot-reload rate limiter
	if cfg.EnableRateLimiting {
		e.limiter.SetRate(cfg.RateLimitPerSecond, cfg.RateLimitBurst)
	}

	// Hot-reload pressure thresholds
	e.pressure.UpdateConfig(PressureConfig{
		HighQueueDepth:     cfg.HighPressureQueueDepth,
		CriticalQueueDepth: cfg.CritPressureQueueDepth,
		HighLatency:        time.Duration(cfg.HighPressureLatencyMs) * time.Millisecond,
		CriticalLatency:    time.Duration(cfg.CritPressureLatencyMs) * time.Millisecond,
	})
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

	log.Printf("[matchmaker] Engine started with %v window (rate limit: %.1f tok/s, burst %d)",
		windowDuration, e.limiter.Rate(), e.limiter.Burst())

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
			e.admitAndRun(ctx)
		case <-e.triggerCh:
			e.admitAndRun(ctx)
		}
	}
}

// admitAndRun applies the rate limiter admission gate before running a round.
func (e *Engine) admitAndRun(ctx context.Context) {
	if !e.limiter.Allow() {
		// Rate limited — drop this round gracefully
		e.mu.Lock()
		e.status.DroppedRounds++
		e.mu.Unlock()
		log.Printf("[matchmaker] ⚠️ Rate limited — round dropped (total dropped: %d)", e.status.DroppedRounds)
		return
	}
	e.runMatchingRound(ctx)
}

// runMatchingRound executes a single matching round with pressure-aware algo selection.
func (e *Engine) runMatchingRound(ctx context.Context) {
	roundStart := time.Now()

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

	// 3. Observe queue depth for pressure monitoring
	queueDepth := len(e.triggerCh)
	tier := e.pressure.Tier()

	// 4. Select algorithm based on pressure tier + circuit breaker
	cbErr := e.breaker.Execute(func() error {
		switch tier {
		case PressureCritical:
			// 🚨 CRITICAL — O(n) Nearest Neighbour, skip everything else
			assignments = NearestNeighborMatch(workers, jobs, cfg, now)
			log.Printf("[matchmaker] Window %d: 🚨 CRITICAL pressure — %d assignments (nearest-neighbor, queue=%d)",
				windowID, len(assignments), queueDepth)

		case PressureHigh:
			// ⚠️ HIGH — fast greedy matching (skip Hungarian + stability)
			assignments = greedyMatch(workers, jobs, cfg, now)
			log.Printf("[matchmaker] Window %d: ⚠️ HIGH pressure — %d assignments (greedy, queue=%d)",
				windowID, len(assignments), queueDepth)

		default:
			// ✅ NORMAL — full accuracy path
			if cfg.EnablePruning && cfg.KNearestNeighbors > 0 {
				// Dynamic Candidate Pruning: use H3 k-ring spatial index
				assignments = PruneAndMatch(workers, jobs, cfg, now)
				log.Printf("[matchmaker] Window %d: %d workers × %d jobs → %d assignments (pruned, k=%d)",
					windowID, len(workers), len(jobs), len(assignments), cfg.KNearestNeighbors)
			} else {
				// Full matrix: classic Hungarian on all W×J pairs
				weightMatrix := BuildWeightMatrix(workers, jobs, cfg, now)
				assignments = SolveMaxWeightMatching(weightMatrix, workers, jobs, cfg.MinWeightThreshold)
				log.Printf("[matchmaker] Window %d: %d workers × %d jobs → %d assignments (full matrix)",
					windowID, len(workers), len(jobs), len(assignments))
			}
		}
		return nil
	})

	if cbErr != nil {
		// Circuit breaker is OPEN — ultimate fallback to greedy
		log.Printf("[matchmaker] Window %d: circuit breaker OPEN, falling back to greedy matching", windowID)
		assignments = greedyMatch(workers, jobs, cfg, now)
		log.Printf("[matchmaker] Window %d: %d assignments (greedy fallback)", windowID, len(assignments))
	}

	// 5. Enforce stability only under NORMAL pressure (§3.1.2)
	if cfg.EnableStability && tier == PressureNormal {
		assignments = EnforceStability(assignments, workers, jobs, cfg, now)
		log.Printf("[matchmaker] Window %d: %d stable assignments after Gale-Shapley refinement",
			windowID, len(assignments))
	}

	// 6. Shadow matching only under NORMAL pressure (A/B testing)
	if cfg.EnableShadowMode && tier == PressureNormal {
		shadowResult := RunShadowMatching(workers, jobs, cfg, now, assignments)
		e.mu.Lock()
		e.status.LastShadowDelta = shadowResult.Comparison.WeightDeltaPct
		e.mu.Unlock()
	}

	// 7. Execute assignments with distributed locking
	for _, a := range assignments {
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

	// 8. Record round latency for pressure monitor
	roundDuration := time.Since(roundStart)
	e.pressure.Record(queueDepth, roundDuration)

	// 9. Publish result
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

	// 10. Update status
	e.mu.Lock()
	e.status.TotalRounds++
	e.status.TotalMatches += int64(len(assignments))
	e.status.LastRunAt = now
	e.status.RoundLatencyMs = roundDuration.Milliseconds()
	e.mu.Unlock()
}
