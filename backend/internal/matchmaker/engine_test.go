package matchmaker

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEngine_RunAndStop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WindowDurationSeconds = 1 // 1-second window for fast test

	fetchWorkers := func(ctx context.Context) ([]MatchableWorker, error) {
		return []MatchableWorker{
			{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5, IsOnline: true},
		}, nil
	}

	fetchJobs := func(ctx context.Context) ([]MatchableJob, error) {
		return []MatchableJob{
			{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: time.Now()},
		}, nil
	}

	var mu sync.Mutex
	var assigned []Assignment
	onAssign := func(ctx context.Context, a Assignment) error {
		mu.Lock()
		assigned = append(assigned, a)
		mu.Unlock()
		return nil
	}

	engine := NewEngine(cfg, fetchWorkers, fetchJobs, onAssign)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go engine.Run(ctx)

	// Wait for at least one round
	time.Sleep(2 * time.Second)
	cancel()

	// Give engine time to shut down
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(assigned) == 0 {
		t.Error("expected at least 1 assignment after running the engine")
	}

	status := engine.Status()
	if status.TotalRounds == 0 {
		t.Error("expected at least 1 round")
	}
	if status.TotalMatches == 0 {
		t.Error("expected at least 1 match")
	}

	t.Logf("Engine ran %d rounds, %d total matches", status.TotalRounds, status.TotalMatches)
}

func TestEngine_EmptyWindow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WindowDurationSeconds = 1

	fetchWorkers := func(ctx context.Context) ([]MatchableWorker, error) {
		return nil, nil // No workers
	}
	fetchJobs := func(ctx context.Context) ([]MatchableJob, error) {
		return nil, nil // No jobs
	}

	var assigned []Assignment
	onAssign := func(ctx context.Context, a Assignment) error {
		assigned = append(assigned, a)
		return nil
	}

	engine := NewEngine(cfg, fetchWorkers, fetchJobs, onAssign)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go engine.Run(ctx)
	time.Sleep(1500 * time.Millisecond)
	cancel()
	time.Sleep(500 * time.Millisecond)

	if len(assigned) != 0 {
		t.Errorf("expected 0 assignments for empty window, got %d", len(assigned))
	}

	status := engine.Status()
	if status.TotalMatches != 0 {
		t.Errorf("expected 0 matches, got %d", status.TotalMatches)
	}
}

func TestEngine_ManualTrigger(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WindowDurationSeconds = 60 // long window, won't fire naturally

	fetchWorkers := func(ctx context.Context) ([]MatchableWorker, error) {
		return []MatchableWorker{
			{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.0, IsOnline: true},
		}, nil
	}
	fetchJobs := func(ctx context.Context) ([]MatchableJob, error) {
		return []MatchableJob{
			{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: time.Now()},
		}, nil
	}

	var mu sync.Mutex
	var assigned []Assignment
	onAssign := func(ctx context.Context, a Assignment) error {
		mu.Lock()
		assigned = append(assigned, a)
		mu.Unlock()
		return nil
	}

	engine := NewEngine(cfg, fetchWorkers, fetchJobs, onAssign)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go engine.Run(ctx)

	// Wait for engine to start, then trigger manually
	time.Sleep(500 * time.Millisecond)
	engine.Trigger()
	time.Sleep(1 * time.Second)

	mu.Lock()
	count := len(assigned)
	mu.Unlock()

	if count == 0 {
		t.Error("expected assignment after manual trigger")
	}

	cancel()
	time.Sleep(500 * time.Millisecond)

	t.Logf("Manual trigger produced %d assignments", count)
}

func TestEngine_ConfigUpdate(t *testing.T) {
	cfg := DefaultConfig()
	engine := NewEngine(cfg, nil, nil, nil)

	newCfg := cfg
	newCfg.AlphaProximity = 0.50
	newCfg.WindowDurationSeconds = 60

	engine.UpdateConfig(newCfg)

	got := engine.Config()
	if got.AlphaProximity != 0.50 {
		t.Errorf("expected AlphaProximity 0.50, got %f", got.AlphaProximity)
	}
	if got.WindowDurationSeconds != 60 {
		t.Errorf("expected WindowDurationSeconds 60, got %d", got.WindowDurationSeconds)
	}
}
