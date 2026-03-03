package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// ErrDuplicateRequest is returned when a request with the same idempotency key
// has already been processed.
var ErrDuplicateRequest = errors.New("idempotency: duplicate request")

// Store provides idempotency key tracking to prevent duplicate operations.
// Every mutating request (CreateJob, AcceptJob, Payment) should carry a unique
// idempotency key. If the same key is seen again, the original result is returned.
//
// In production, swap InMemoryStore for a Redis-backed implementation.
type Store interface {
	// Check returns (true, cachedResult) if the key was already processed,
	// or (false, nil) if this is a new request.
	Check(ctx context.Context, key string) (exists bool, result []byte, err error)

	// Set stores the result for an idempotency key with a TTL.
	// After TTL expires, the key can be reused.
	Set(ctx context.Context, key string, result []byte, ttl time.Duration) error
}

// --- In-Memory Implementation ---

type entry struct {
	result    []byte
	expiresAt time.Time
}

// InMemoryStore is a single-process idempotency store.
// Suitable for development, testing, and single-instance deployments.
type InMemoryStore struct {
	mu      sync.RWMutex
	entries map[string]entry
}

// NewInMemoryStore creates a new in-memory idempotency store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		entries: make(map[string]entry),
	}
}

func (s *InMemoryStore) Check(ctx context.Context, key string) (bool, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.entries[key]
	if !ok {
		return false, nil, nil
	}

	// Check expiry
	if time.Now().After(e.expiresAt) {
		return false, nil, nil
	}

	return true, e.result, nil
}

func (s *InMemoryStore) Set(ctx context.Context, key string, result []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = entry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// --- Helper Functions ---

// MarshalResult marshals a result to JSON for storage.
func MarshalResult(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalResult unmarshals a stored result from JSON.
func UnmarshalResult(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
