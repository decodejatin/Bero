package idempotency

import (
	"context"
	"testing"
	"time"
)

func TestCheckAndSet(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// First check — should not exist
	exists, _, err := store.Check(ctx, "order-123")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if exists {
		t.Fatal("key should not exist yet")
	}

	// Set result
	result, _ := MarshalResult(map[string]string{"job_id": "j1"})
	if err := store.Set(ctx, "order-123", result, 5*time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	// Second check — should exist
	exists, cached, err := store.Check(ctx, "order-123")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if !exists {
		t.Fatal("key should exist after set")
	}

	var decoded map[string]string
	if err := UnmarshalResult(cached, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded["job_id"] != "j1" {
		t.Errorf("expected job_id j1, got %s", decoded["job_id"])
	}
}

func TestTTLExpiry(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	result, _ := MarshalResult("test")
	store.Set(ctx, "key-expire", result, 50*time.Millisecond)

	// Should exist immediately
	exists, _, _ := store.Check(ctx, "key-expire")
	if !exists {
		t.Fatal("should exist immediately after set")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should be gone after TTL
	exists, _, _ = store.Check(ctx, "key-expire")
	if exists {
		t.Fatal("should not exist after TTL expiry")
	}
}

func TestDifferentKeys(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	r1, _ := MarshalResult("result1")
	r2, _ := MarshalResult("result2")
	store.Set(ctx, "key-a", r1, 5*time.Minute)
	store.Set(ctx, "key-b", r2, 5*time.Minute)

	exists1, data1, _ := store.Check(ctx, "key-a")
	exists2, data2, _ := store.Check(ctx, "key-b")
	exists3, _, _ := store.Check(ctx, "key-c")

	if !exists1 || !exists2 || exists3 {
		t.Error("key existence mismatch")
	}

	var s1, s2 string
	UnmarshalResult(data1, &s1)
	UnmarshalResult(data2, &s2)
	if s1 != "result1" || s2 != "result2" {
		t.Errorf("data mismatch: %s, %s", s1, s2)
	}
}
