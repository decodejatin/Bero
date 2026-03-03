package matchmaker

import (
	"testing"
	"time"
)

func TestPressureTier_String(t *testing.T) {
	cases := map[PressureTier]string{
		PressureNormal:   "NORMAL",
		PressureHigh:     "HIGH",
		PressureCritical: "CRITICAL",
	}
	for tier, want := range cases {
		if got := tier.String(); got != want {
			t.Errorf("PressureTier(%d).String() = %q, want %q", tier, got, want)
		}
	}
}

func TestPressureMonitor_DefaultIsNormal(t *testing.T) {
	m := NewPressureMonitor(DefaultPressureConfig())
	if m.Tier() != PressureNormal {
		t.Errorf("new monitor should start at NORMAL, got %s", m.Tier())
	}
}

func TestPressureMonitor_QueueDepthEscalation(t *testing.T) {
	cfg := DefaultPressureConfig()
	m := NewPressureMonitor(cfg)

	// Record with queue depth just below HIGH threshold
	m.Record(cfg.HighQueueDepth-1, 10*time.Millisecond)
	if m.Tier() != PressureNormal {
		t.Errorf("below HIGH threshold: expected NORMAL, got %s", m.Tier())
	}

	// Record with queue depth at HIGH threshold
	m.Record(cfg.HighQueueDepth, 10*time.Millisecond)
	if m.Tier() != PressureHigh {
		t.Errorf("at HIGH threshold: expected HIGH, got %s", m.Tier())
	}

	// Record with queue depth at CRITICAL threshold
	m.Record(cfg.CriticalQueueDepth, 10*time.Millisecond)
	if m.Tier() != PressureCritical {
		t.Errorf("at CRITICAL threshold: expected CRITICAL, got %s", m.Tier())
	}
}

func TestPressureMonitor_LatencyEscalation(t *testing.T) {
	cfg := DefaultPressureConfig() // HIGH=500ms, CRITICAL=2s
	m := NewPressureMonitor(cfg)

	// Seed with a fast round first (seed EWMA)
	m.Record(0, 10*time.Millisecond)
	if m.Tier() != PressureNormal {
		t.Errorf("fast round: expected NORMAL, got %s", m.Tier())
	}

	// Feed many slow rounds — EWMA with α=0.3 needs several samples
	// to cross the 500ms threshold when seeded at 10ms.
	for i := 0; i < 10; i++ {
		m.Record(0, cfg.HighLatency+200*time.Millisecond)
	}
	if m.Tier() != PressureHigh {
		t.Errorf("sustained slow rounds: expected HIGH, got %s", m.Tier())
	}

	// Critical latency rounds
	for i := 0; i < 10; i++ {
		m.Record(0, cfg.CriticalLatency+500*time.Millisecond)
	}
	if m.Tier() != PressureCritical {
		t.Errorf("critical latency rounds: expected CRITICAL, got %s", m.Tier())
	}
}

func TestPressureMonitor_RecoveryAfterHighLoad(t *testing.T) {
	cfg := DefaultPressureConfig()
	m := NewPressureMonitor(cfg)

	// Escalate to CRITICAL via queue depth
	m.Record(cfg.CriticalQueueDepth+10, 10*time.Millisecond)
	if m.Tier() != PressureCritical {
		t.Fatalf("expected CRITICAL after backlog, got %s", m.Tier())
	}

	// Queue drains — many fast rounds with empty queues
	for i := 0; i < 20; i++ {
		m.Record(0, 5*time.Millisecond)
	}
	if m.Tier() != PressureNormal {
		t.Errorf("expected NORMAL after queue drains, got %s", m.Tier())
	}
}

func TestPressureMonitor_Stats(t *testing.T) {
	m := NewPressureMonitor(DefaultPressureConfig())
	m.Record(7, 300*time.Millisecond)

	tier, ewma, depth := m.Stats()
	if tier != PressureNormal {
		t.Errorf("expected NORMAL, got %s", tier)
	}
	if ewma == 0 {
		t.Error("EWMA latency should be non-zero after Record()")
	}
	if depth != 7 {
		t.Errorf("expected queue depth 7, got %d", depth)
	}
}

func TestPressureMonitor_UpdateConfig(t *testing.T) {
	m := NewPressureMonitor(DefaultPressureConfig())
	m.Record(5, 100*time.Millisecond)
	if m.Tier() != PressureNormal {
		t.Fatal("expected NORMAL")
	}

	// Lower HIGH threshold so depth=5 now triggers it
	m.UpdateConfig(PressureConfig{
		HighQueueDepth:     3,
		CriticalQueueDepth: 100,
		HighLatency:        10 * time.Second,
		CriticalLatency:    60 * time.Second,
	})
	if m.Tier() != PressureHigh {
		t.Errorf("after config update lowering threshold, expected HIGH, got %s", m.Tier())
	}
}
