package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryLimiter_BurstAllowed(t *testing.T) {
	// Burst of 5 should allow 5 rapid requests
	lim := NewInMemoryLimiter(1.0, 5)
	for i := 0; i < 5; i++ {
		if !lim.Allow() {
			t.Errorf("request %d should be allowed within burst capacity", i+1)
		}
	}
	// 6th should be denied (bucket empty, refill too slow at 1/s)
	if lim.Allow() {
		t.Errorf("6th request should be denied — burst exceeded at 1 token/s")
	}
}

func TestInMemoryLimiter_SustainedRateThrottled(t *testing.T) {
	// Rate of 100/s, burst 1 — only 1 immediate then throttled
	lim := NewInMemoryLimiter(100.0, 1)
	if !lim.Allow() {
		t.Fatal("first request should be allowed")
	}
	// Second immediate request should be denied (burst=1 exhausted)
	if lim.Allow() {
		t.Error("second immediate request should be throttled")
	}
	// After 15ms at 100/s we get ~1.5 tokens back
	time.Sleep(15 * time.Millisecond)
	if !lim.Allow() {
		t.Error("request after refill period should be allowed")
	}
}

func TestInMemoryLimiter_SetRateHotReload(t *testing.T) {
	lim := NewInMemoryLimiter(1.0, 1)

	// Exhaust the bucket
	lim.Allow()
	if lim.Allow() {
		t.Fatal("should be exhausted")
	}

	// Hot-reload to much higher rate
	lim.SetRate(1000.0, 10)

	// Verify getters reflect new config
	if lim.Rate() != 1000.0 {
		t.Errorf("expected rate 1000.0, got %.1f", lim.Rate())
	}
	if lim.Burst() != 10 {
		t.Errorf("expected burst 10, got %d", lim.Burst())
	}

	// At 1000 tok/s, wait just 5ms for tokens to refill
	time.Sleep(5 * time.Millisecond)
	if !lim.Allow() {
		t.Error("after SetRate to 1000/s and 5ms sleep, token should be available")
	}
}

func TestInMemoryLimiter_Wait(t *testing.T) {
	// At 10 tokens/sec, Wait should complete in < 200ms
	lim := NewInMemoryLimiter(10.0, 1)
	lim.Allow() // exhaust

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := lim.Wait(ctx); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Wait took too long: %v (expected <200ms at 10 tok/s)", elapsed)
	}
}

func TestInMemoryLimiter_WaitCancelled(t *testing.T) {
	// Rate 0.001/s — will never refill in test time
	lim := NewInMemoryLimiter(0.001, 1)
	lim.Allow() // exhaust

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := lim.Wait(ctx); err == nil {
		t.Error("Wait should return error when context is cancelled")
	}
}

func TestNoopLimiter_AlwaysAllow(t *testing.T) {
	lim := NewNoopLimiter()
	for i := 0; i < 1000; i++ {
		if !lim.Allow() {
			t.Fatalf("NoopLimiter must always allow, failed at iteration %d", i)
		}
	}
}

func TestInMemoryLimiter_RateAndBurst(t *testing.T) {
	lim := NewInMemoryLimiter(5.0, 10)
	if lim.Rate() != 5.0 {
		t.Errorf("expected rate 5.0, got %.1f", lim.Rate())
	}
	if lim.Burst() != 10 {
		t.Errorf("expected burst 10, got %d", lim.Burst())
	}
	lim.SetRate(2.0, 3)
	if lim.Rate() != 2.0 {
		t.Errorf("expected rate 2.0 after SetRate, got %.1f", lim.Rate())
	}
	if lim.Burst() != 3 {
		t.Errorf("expected burst 3 after SetRate, got %d", lim.Burst())
	}
}
