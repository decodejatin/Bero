package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

var errSimulated = errors.New("simulated failure")

func TestClosedState_Success(t *testing.T) {
	cb := New(DefaultConfig())

	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("expected CLOSED, got %s", cb.State())
	}
}

func TestClosedState_FailuresToOpen(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxFailures = 3
	cb := New(cfg)

	// 2 failures — still closed
	for i := 0; i < 2; i++ {
		cb.Execute(func() error { return errSimulated })
	}
	if cb.State() != StateClosed {
		t.Errorf("should be CLOSED after %d failures, got %s", 2, cb.State())
	}

	// 3rd failure — opens
	cb.Execute(func() error { return errSimulated })
	if cb.State() != StateOpen {
		t.Errorf("should be OPEN after %d failures, got %s", 3, cb.State())
	}
}

func TestOpenState_RejectsAll(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxFailures = 1
	cfg.ResetTimeout = 1 * time.Hour // Won't expire during test
	cb := New(cfg)

	// Trip the breaker
	cb.Execute(func() error { return errSimulated })

	// All calls should be rejected
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestOpenToHalfOpen(t *testing.T) {
	cfg := Config{
		MaxFailures:  1,
		ResetTimeout: 50 * time.Millisecond,
		HalfOpenMax:  1,
	}
	cb := New(cfg)

	// Trip the breaker
	cb.Execute(func() error { return errSimulated })
	if cb.State() != StateOpen {
		t.Fatalf("expected OPEN, got %s", cb.State())
	}

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Next call should go through (half-open)
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("half-open call should succeed: %v", err)
	}

	// Should be closed now (1 success in half-open)
	if cb.State() != StateClosed {
		t.Errorf("should be CLOSED after half-open success, got %s", cb.State())
	}
}

func TestHalfOpen_FailureGoesBackToOpen(t *testing.T) {
	cfg := Config{
		MaxFailures:  1,
		ResetTimeout: 50 * time.Millisecond,
		HalfOpenMax:  1,
	}
	cb := New(cfg)

	// Trip the breaker
	cb.Execute(func() error { return errSimulated })

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Fail the half-open call
	cb.Execute(func() error { return errSimulated })

	if cb.State() != StateOpen {
		t.Errorf("should be OPEN after half-open failure, got %s", cb.State())
	}
}

func TestSuccessResetsFailureCount(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxFailures = 3
	cb := New(cfg)

	// 2 failures
	cb.Execute(func() error { return errSimulated })
	cb.Execute(func() error { return errSimulated })

	// 1 success — resets count
	cb.Execute(func() error { return nil })

	// 2 more failures — should NOT open (count was reset)
	cb.Execute(func() error { return errSimulated })
	cb.Execute(func() error { return errSimulated })

	if cb.State() != StateClosed {
		t.Errorf("should still be CLOSED after reset, got %s", cb.State())
	}
}

func TestManualReset(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxFailures = 1
	cb := New(cfg)

	cb.Execute(func() error { return errSimulated })
	if cb.State() != StateOpen {
		t.Fatal("should be OPEN")
	}

	cb.Reset()
	if cb.State() != StateClosed {
		t.Errorf("should be CLOSED after reset, got %s", cb.State())
	}

	// Should work normally now
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Errorf("should succeed after reset: %v", err)
	}
}
