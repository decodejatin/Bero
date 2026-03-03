package eventbus

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event published to the event bus.
type Event struct {
	ID        string      `json:"id"`        // unique event ID (idempotency)
	Type      string      `json:"type"`      // e.g. "order.placed", "worker.assigned"
	Payload   interface{} `json:"payload"`   // event-specific data
	Timestamp time.Time   `json:"timestamp"` // when the event was created
}

// NewEvent creates a new event with auto-generated ID and timestamp.
func NewEvent(eventType string, payload interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// Well-known event types for the Bero platform.
const (
	EventOrderPlaced    = "order.placed"
	EventWorkerAssigned = "worker.assigned"
	EventJobStarted     = "job.started"
	EventJobCompleted   = "job.completed"
	EventJobCancelled   = "job.cancelled"
)

// --- Event Payloads ---

// OrderPlacedPayload carries data for order.placed events.
type OrderPlacedPayload struct {
	JobID    string  `json:"job_id"`
	ClientID string  `json:"client_id"`
	Category string  `json:"category"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	IsUrgent bool    `json:"is_urgent"`
}

// WorkerAssignedPayload carries data for worker.assigned events.
type WorkerAssignedPayload struct {
	JobID    string  `json:"job_id"`
	WorkerID string  `json:"worker_id"`
	Weight   float64 `json:"weight"`
}

// JobLifecyclePayload carries data for job lifecycle events.
type JobLifecyclePayload struct {
	JobID    string `json:"job_id"`
	WorkerID string `json:"worker_id,omitempty"`
	Status   string `json:"status"`
}

// Handler is a function that processes an event.
type Handler func(event Event)

// EventBus provides publish/subscribe messaging for domain events.
// In production, swap InMemoryEventBus for a NATS implementation.
type EventBus interface {
	// Publish sends an event to all subscribers of that event type.
	Publish(ctx context.Context, event Event) error

	// Subscribe registers a handler for a specific event type.
	// Returns an unsubscribe function.
	Subscribe(eventType string, handler Handler) func()
}

// --- In-Memory Implementation ---

// InMemoryEventBus is a channel-based event bus for single-instance deployments.
// Suitable for development, testing, and early-stage production.
// For multi-instance production, replace with NATS.
type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]subscriberEntry
	nextID      int
}

type subscriberEntry struct {
	id      int
	handler Handler
}

// NewInMemoryEventBus creates a new in-memory event bus.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[string][]subscriberEntry),
	}
}

func (b *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	subs := b.subscribers[event.Type]
	// Copy to avoid holding lock during handler execution
	handlers := make([]Handler, len(subs))
	for i, s := range subs {
		handlers[i] = s.handler
	}
	b.mu.RUnlock()

	// Fire-and-forget: handlers run asynchronously
	for _, h := range handlers {
		go func(handler Handler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[eventbus] handler panic for %s: %v", event.Type, r)
				}
			}()
			handler(event)
		}(h)
	}

	return nil
}

func (b *InMemoryEventBus) Subscribe(eventType string, handler Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID

	b.subscribers[eventType] = append(b.subscribers[eventType], subscriberEntry{
		id:      id,
		handler: handler,
	})

	// Return unsubscribe function
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[eventType]
		for i, s := range subs {
			if s.id == id {
				b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}
