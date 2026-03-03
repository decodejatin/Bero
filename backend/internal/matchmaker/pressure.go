// Package matchmaker — pressure.go
// PressureMonitor tracks real-time engine load and derives a PressureTier
// that drives algorithm selection (Hungarian → Greedy → Nearest Neighbour).
//
// The monitor uses an Exponentially Weighted Moving Average (EWMA) for latency
// to smooth out spikes and avoid thrashing between tiers.
package matchmaker

import (
	"sync"
	"time"
)

// PressureTier represents the current load level of the matching engine.
type PressureTier int

const (
	// PressureNormal — engine is healthy. Run full Hungarian / PruneAndMatch.
	PressureNormal PressureTier = iota

	// PressureHigh — queue is building up. Switch to fast Greedy matching.
	// Accepts ~90% match quality at ~10× the speed.
	PressureHigh

	// PressureCritical — backlog is severe. Switch to O(n) Nearest Neighbour.
	// Clears the queue at maximum speed regardless of optimality.
	PressureCritical
)

// String returns a human-readable tier name for logging and status APIs.
func (p PressureTier) String() string {
	switch p {
	case PressureHigh:
		return "HIGH"
	case PressureCritical:
		return "CRITICAL"
	default:
		return "NORMAL"
	}
}

// PressureConfig holds thresholds for tier transitions.
type PressureConfig struct {
	// Queue depth thresholds — how many unprocessed triggers are waiting
	HighQueueDepth     int // default 10
	CriticalQueueDepth int // default 50

	// Round latency thresholds (EWMA)
	HighLatency     time.Duration // default 500ms
	CriticalLatency time.Duration // default 2s
}

// DefaultPressureConfig returns sensible production defaults.
func DefaultPressureConfig() PressureConfig {
	return PressureConfig{
		HighQueueDepth:     10,
		CriticalQueueDepth: 50,
		HighLatency:        500 * time.Millisecond,
		CriticalLatency:    2 * time.Second,
	}
}

// PressureMonitor tracks queue depth and EWMA latency to derive PressureTier.
// All methods are safe for concurrent use.
type PressureMonitor struct {
	mu  sync.RWMutex
	cfg PressureConfig

	ewmaLatency    time.Duration // current EWMA latency estimate
	ewmaAlpha      float64       // EWMA smoothing factor (0 < α ≤ 1)
	lastQueueDepth int
	currentTier    PressureTier
	roundCount     int64
}

// NewPressureMonitor creates a monitor with the given config.
func NewPressureMonitor(cfg PressureConfig) *PressureMonitor {
	return &PressureMonitor{
		cfg:       cfg,
		ewmaAlpha: 0.3, // weights recent samples more; larger = faster response
	}
}

// Record updates the monitor with the latest observation after a matching round.
//   - queueDepth: number of pending trigger signals in the engine's trigger channel
//   - roundDuration: wall-clock time the matching round took end-to-end
func (m *PressureMonitor) Record(queueDepth int, roundDuration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.roundCount++
	m.lastQueueDepth = queueDepth

	// EWMA update: ewma = α·sample + (1-α)·ewma
	if m.roundCount == 1 {
		m.ewmaLatency = roundDuration // seed with first observation
	} else {
		alpha := m.ewmaAlpha
		m.ewmaLatency = time.Duration(alpha*float64(roundDuration) + (1-alpha)*float64(m.ewmaLatency))
	}

	// Compute new tier
	m.currentTier = m.computeTier()
}

// Tier returns the current pressure tier based on latest observations.
func (m *PressureMonitor) Tier() PressureTier {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTier
}

// Stats returns current monitor readings for status endpoints.
func (m *PressureMonitor) Stats() (tier PressureTier, ewmaLatency time.Duration, queueDepth int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTier, m.ewmaLatency, m.lastQueueDepth
}

// UpdateConfig hot-reloads thresholds at runtime.
func (m *PressureMonitor) UpdateConfig(cfg PressureConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = cfg
	m.currentTier = m.computeTier()
}

// computeTier evaluates current readings against thresholds.
// Must be called with m.mu held (write or read+recompute).
func (m *PressureMonitor) computeTier() PressureTier {
	// Critical if EITHER latency OR queue depth exceeds critical threshold
	if m.lastQueueDepth >= m.cfg.CriticalQueueDepth ||
		(m.roundCount > 0 && m.ewmaLatency >= m.cfg.CriticalLatency) {
		return PressureCritical
	}
	// High if EITHER latency OR queue depth exceeds high threshold
	if m.lastQueueDepth >= m.cfg.HighQueueDepth ||
		(m.roundCount > 0 && m.ewmaLatency >= m.cfg.HighLatency) {
		return PressureHigh
	}
	return PressureNormal
}
