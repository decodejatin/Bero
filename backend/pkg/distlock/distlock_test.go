package distlock

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAcquireAndRelease(t *testing.T) {
	dl := NewInMemoryDistLock()
	ctx := context.Background()

	lock, err := dl.Acquire(ctx, "worker:w1", 5*time.Second)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Same key should fail
	_, err = dl.Acquire(ctx, "worker:w1", 5*time.Second)
	if err != ErrLockNotAcquired {
		t.Fatalf("expected ErrLockNotAcquired, got %v", err)
	}

	// Different key should succeed
	lock2, err := dl.Acquire(ctx, "worker:w2", 5*time.Second)
	if err != nil {
		t.Fatalf("different key should succeed: %v", err)
	}

	// Release first lock
	if err := lock.Release(ctx); err != nil {
		t.Fatalf("failed to release: %v", err)
	}

	// Now same key should succeed
	lock3, err := dl.Acquire(ctx, "worker:w1", 5*time.Second)
	if err != nil {
		t.Fatalf("after release should succeed: %v", err)
	}

	// Double release should error
	if err := lock.Release(ctx); err == nil {
		t.Error("double release should error")
	}

	// Cleanup
	lock2.Release(ctx)
	lock3.Release(ctx)
}

func TestLockTTLExpiry(t *testing.T) {
	dl := NewInMemoryDistLock()
	ctx := context.Background()

	_, err := dl.Acquire(ctx, "worker:w1", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to acquire: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should succeed after expiry
	lock, err := dl.Acquire(ctx, "worker:w1", 5*time.Second)
	if err != nil {
		t.Fatalf("should acquire after TTL expiry: %v", err)
	}
	lock.Release(ctx)
}

func TestConcurrentAcquisition(t *testing.T) {
	dl := NewInMemoryDistLock()
	ctx := context.Background()

	const goroutines = 50
	var acquired int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			lock, err := dl.Acquire(ctx, "worker:contested", 5*time.Second)
			if err == nil {
				mu.Lock()
				acquired++
				mu.Unlock()
				// Hold briefly then release
				time.Sleep(10 * time.Millisecond)
				lock.Release(ctx)
			}
		}()
	}

	wg.Wait()

	// Exactly 1 goroutine should have acquired (others fail fast)
	if acquired != 1 {
		t.Errorf("expected exactly 1 acquisition, got %d", acquired)
	}
}
