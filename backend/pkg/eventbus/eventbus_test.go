package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	var received Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe(EventOrderPlaced, func(e Event) {
		received = e
		wg.Done()
	})

	event := NewEvent(EventOrderPlaced, OrderPlacedPayload{
		JobID:    "j1",
		ClientID: "c1",
		Category: "PLUMBING",
	})

	bus.Publish(ctx, event)

	// Wait for async handler
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event handler")
	}

	if received.ID != event.ID {
		t.Errorf("expected event ID %s, got %s", event.ID, received.ID)
	}
	if received.Type != EventOrderPlaced {
		t.Errorf("expected type %s, got %s", EventOrderPlaced, received.Type)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	var mu sync.Mutex
	count := 0
	var wg sync.WaitGroup
	wg.Add(3)

	for i := 0; i < 3; i++ {
		bus.Subscribe(EventWorkerAssigned, func(e Event) {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
		})
	}

	bus.Publish(ctx, NewEvent(EventWorkerAssigned, WorkerAssignedPayload{
		JobID:    "j1",
		WorkerID: "w1",
		Weight:   0.85,
	}))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	mu.Lock()
	if count != 3 {
		t.Errorf("expected 3 handlers called, got %d", count)
	}
	mu.Unlock()
}

func TestUnsubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	var mu sync.Mutex
	count := 0

	unsub := bus.Subscribe(EventJobCompleted, func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	// Publish once
	bus.Publish(ctx, NewEvent(EventJobCompleted, nil))
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	c1 := count
	mu.Unlock()

	// Unsubscribe
	unsub()

	// Publish again — should not increment
	bus.Publish(ctx, NewEvent(EventJobCompleted, nil))
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	c2 := count
	mu.Unlock()

	if c1 != 1 {
		t.Errorf("expected 1 call before unsub, got %d", c1)
	}
	if c2 != 1 {
		t.Errorf("expected still 1 call after unsub, got %d", c2)
	}
}

func TestNoSubscribers(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	// Should not panic
	err := bus.Publish(ctx, NewEvent(EventJobCancelled, nil))
	if err != nil {
		t.Errorf("publish with no subscribers should not error: %v", err)
	}
}

func TestEventTypeIsolation(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	var mu sync.Mutex
	placedCount := 0
	assignedCount := 0

	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe(EventOrderPlaced, func(e Event) {
		mu.Lock()
		placedCount++
		mu.Unlock()
	})
	bus.Subscribe(EventWorkerAssigned, func(e Event) {
		mu.Lock()
		assignedCount++
		mu.Unlock()
		wg.Done()
	})

	// Only publish worker.assigned
	bus.Publish(ctx, NewEvent(EventWorkerAssigned, nil))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	time.Sleep(50 * time.Millisecond) // Extra time for potential cross-fire

	mu.Lock()
	if placedCount != 0 {
		t.Errorf("order.placed handler should not fire, count=%d", placedCount)
	}
	if assignedCount != 1 {
		t.Errorf("worker.assigned handler should fire once, count=%d", assignedCount)
	}
	mu.Unlock()
}
