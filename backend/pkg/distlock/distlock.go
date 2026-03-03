package distlock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrLockNotAcquired is returned when a lock cannot be obtained (already held).
var ErrLockNotAcquired = errors.New("distlock: lock not acquired")

// Lock represents an acquired distributed lock that must be released.
type Lock interface {
	// Release releases the lock. Returns error if lock was already released or expired.
	Release(ctx context.Context) error
}

// DistLock provides distributed locking to prevent double-booking.
// In production, swap InMemoryDistLock for a Redis Redlock implementation.
type DistLock interface {
	// Acquire attempts to acquire a lock on the given key with a TTL.
	// Returns ErrLockNotAcquired if the lock is already held.
	Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, error)
}

// --- In-Memory Implementation (single-instance, tests) ---

type inMemoryLock struct {
	mu       *sync.Mutex
	store    *InMemoryDistLock
	key      string
	token    string
	released bool
}

func (l *inMemoryLock) Release(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.released {
		return errors.New("distlock: lock already released")
	}

	l.store.mu.Lock()
	defer l.store.mu.Unlock()

	// Only release if our token still holds the lock (wasn't expired)
	if entry, ok := l.store.locks[l.key]; ok && entry.token == l.token {
		delete(l.store.locks, l.key)
	}
	l.released = true
	return nil
}

type lockEntry struct {
	token     string
	expiresAt time.Time
}

// InMemoryDistLock is a single-process distributed lock using sync.Mutex.
// Suitable for development, testing, and single-instance deployments.
// For multi-instance production, replace with Redis Redlock.
type InMemoryDistLock struct {
	mu    sync.Mutex
	locks map[string]lockEntry
}

// NewInMemoryDistLock creates a new in-memory distributed lock.
func NewInMemoryDistLock() *InMemoryDistLock {
	return &InMemoryDistLock{
		locks: make(map[string]lockEntry),
	}
}

func (d *InMemoryDistLock) Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Check if lock exists and hasn't expired
	if entry, ok := d.locks[key]; ok {
		if now.Before(entry.expiresAt) {
			return nil, ErrLockNotAcquired
		}
		// Expired — allow acquisition
		delete(d.locks, key)
	}

	token := uuid.New().String()
	d.locks[key] = lockEntry{
		token:     token,
		expiresAt: now.Add(ttl),
	}

	lockMu := &sync.Mutex{}
	return &inMemoryLock{
		mu:    lockMu,
		store: d,
		key:   key,
		token: token,
	}, nil
}
