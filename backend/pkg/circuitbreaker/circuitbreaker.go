package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open and rejecting calls.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Failing, rejecting all calls
	StateHalfOpen              // Testing recovery with limited calls
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config holds circuit breaker configuration.
type Config struct {
	MaxFailures  int           // failures before opening (default 5)
	ResetTimeout time.Duration // time before trying half-open (default 30s)
	HalfOpenMax  int           // max calls allowed in half-open (default 1)
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
		HalfOpenMax:  1,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
//
// States:
//   - CLOSED: Normal operation. Failures are counted. Opens after MaxFailures.
//   - OPEN: All calls are rejected with ErrCircuitOpen. After ResetTimeout, transitions to HALF_OPEN.
//   - HALF_OPEN: A limited number of calls are allowed through. If they succeed, transitions to CLOSED.
//     If any fail, transitions back to OPEN.
type CircuitBreaker struct {
	mu sync.Mutex

	config       Config
	state        State
	failureCount int
	successCount int // successes in half-open state
	lastFailure  time.Time
}

// New creates a new circuit breaker with the given config.
func New(cfg Config) *CircuitBreaker {
	return &CircuitBreaker{
		config: cfg,
		state:  StateClosed,
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Execute runs the given function through the circuit breaker.
// Returns ErrCircuitOpen if the circuit is open, otherwise returns the function's error.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case StateOpen:
		// Check if enough time has passed to try half-open
		if time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			return cb.executeHalfOpen(fn)
		}
		cb.mu.Unlock()
		return ErrCircuitOpen

	case StateHalfOpen:
		cb.mu.Unlock()
		return cb.executeHalfOpen(fn)

	default: // Closed
		cb.mu.Unlock()
		return cb.executeClosed(fn)
	}
}

func (cb *CircuitBreaker) executeClosed(fn func() error) error {
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()
		if cb.failureCount >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Reset failure count on success
	cb.failureCount = 0
	return nil
}

func (cb *CircuitBreaker) executeHalfOpen(fn func() error) error {
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Failed in half-open — back to open
		cb.state = StateOpen
		cb.lastFailure = time.Now()
		cb.failureCount = cb.config.MaxFailures
		return err
	}

	// Success in half-open
	cb.successCount++
	if cb.successCount >= cb.config.HalfOpenMax {
		// Enough successes — close the circuit
		cb.state = StateClosed
		cb.failureCount = 0
	}
	return nil
}

// Reset manually resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}
