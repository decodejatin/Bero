// Package ratelimiter provides a Token Bucket rate limiter for the Bero matching engine.
//
// The token bucket algorithm works as follows:
//   - A bucket holds up to Burst tokens
//   - Tokens refill at Rate tokens per second
//   - Each request consumes one token; if the bucket is empty, the request is rejected or waits
//
// This is used as the admission gate before the matching engine runs a round.
// Under normal load the bucket stays full. During festival-level surges, tokens
// are exhausted and the engine drops redundant trigger signals gracefully.
package ratelimiter

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// Limiter is the rate limiting interface. Implementations must be safe for concurrent use.
type Limiter interface {
	// Allow reports whether a single event can proceed right now.
	// Non-blocking: returns false if no tokens are available.
	Allow() bool

	// Wait blocks until a token is available or ctx is cancelled.
	// Use this for critical requests that must not be dropped.
	Wait(ctx context.Context) error

	// SetRate hot-reloads the token refill rate and burst capacity.
	// Safe to call at runtime (e.g., from a config update).
	SetRate(tokensPerSecond float64, burst int)

	// Rate returns the current configured rate (tokens/sec).
	Rate() float64

	// Burst returns the current burst capacity.
	Burst() int
}

// InMemoryLimiter is a token bucket Limiter backed by golang.org/x/time/rate.
// It is the production implementation used by the matching engine.
type InMemoryLimiter struct {
	mu      sync.RWMutex
	limiter *rate.Limiter
	rps     float64
	burst   int
}

// NewInMemoryLimiter creates a token bucket rate limiter.
//   - tokensPerSecond: refill rate (e.g. 2.0 = two matching rounds per second max)
//   - burst: maximum burst capacity (e.g. 5 = absorb a short spike of 5 triggers)
func NewInMemoryLimiter(tokensPerSecond float64, burst int) *InMemoryLimiter {
	return &InMemoryLimiter{
		limiter: rate.NewLimiter(rate.Limit(tokensPerSecond), burst),
		rps:     tokensPerSecond,
		burst:   burst,
	}
}

// Allow reports whether a matching round can proceed right now.
func (l *InMemoryLimiter) Allow() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.limiter.Allow()
}

// Wait blocks until a token is available or the context is cancelled.
func (l *InMemoryLimiter) Wait(ctx context.Context) error {
	l.mu.RLock()
	lim := l.limiter
	l.mu.RUnlock()
	return lim.Wait(ctx)
}

// SetRate hot-reloads the rate and burst at runtime.
// This is safe to call while the engine is running (e.g., on festival mode activation).
func (l *InMemoryLimiter) SetRate(tokensPerSecond float64, burst int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rps = tokensPerSecond
	l.burst = burst
	l.limiter.SetLimit(rate.Limit(tokensPerSecond))
	l.limiter.SetBurst(burst)
}

// Rate returns the current configured rate.
func (l *InMemoryLimiter) Rate() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.rps
}

// Burst returns the current burst capacity.
func (l *InMemoryLimiter) Burst() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.burst
}

// NoopLimiter always allows every request. Used when rate limiting is disabled
// (EnableRateLimiting = false) or in unit tests that don't want rate throttling.
type NoopLimiter struct{}

func (NoopLimiter) Allow() bool                  { return true }
func (NoopLimiter) Wait(_ context.Context) error { return nil }
func (NoopLimiter) SetRate(_ float64, _ int)     {}
func (NoopLimiter) Rate() float64                { return 0 }
func (NoopLimiter) Burst() int                   { return 0 }

// NewNoopLimiter returns a Limiter that never throttles.
func NewNoopLimiter() Limiter { return NoopLimiter{} }
