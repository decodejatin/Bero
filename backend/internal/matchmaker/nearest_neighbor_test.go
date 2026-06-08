package matchmaker

import (
	"testing"
	"time"
)

func TestNearestNeighborMatch_Basic(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6100, Longitude: 77.2000, IsOnline: true, RatingAvg: 4.5, TrustScore: 0.85},
		{ID: "w2", Latitude: 28.6300, Longitude: 77.2100, IsOnline: true, RatingAvg: 4.0, TrustScore: 0.75},
		{ID: "w3", Latitude: 28.7000, Longitude: 77.3000, IsOnline: true, RatingAvg: 3.8, TrustScore: 0.65},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6110, Longitude: 77.2010, RequiredSkills: []string{}, CreatedAt: now.Add(-5 * time.Minute)},
		{ID: "j2", Latitude: 28.6290, Longitude: 77.2090, RequiredSkills: []string{}, CreatedAt: now.Add(-3 * time.Minute)},
	}

	assignments := NearestNeighborMatch(workers, jobs, cfg, now)

	if len(assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(assignments))
	}

	// w1 is nearest to j1, w2 is nearest to j2
	// Check that assigned workers are distinct (no double-booking)
	workerIDs := map[string]bool{}
	for _, a := range assignments {
		if workerIDs[a.WorkerID] {
			t.Errorf("worker %s assigned more than once (double-booking)", a.WorkerID)
		}
		workerIDs[a.WorkerID] = true
		if a.Weight < cfg.MinWeightThreshold {
			t.Errorf("assignment weight %.4f below threshold %.4f", a.Weight, cfg.MinWeightThreshold)
		}
	}
}

func TestNearestNeighborMatch_EmptyInputs(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	if got := NearestNeighborMatch(nil, nil, cfg, now); got != nil {
		t.Errorf("nil inputs: expected nil, got %v", got)
	}
	if got := NearestNeighborMatch([]MatchableWorker{{ID: "w1"}}, nil, cfg, now); got != nil {
		t.Errorf("nil jobs: expected nil, got %v", got)
	}
}

func TestNearestNeighborMatch_TooFarWorker(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxDistanceKm = 1.0 // very tight radius
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6100, Longitude: 77.2000, IsOnline: true, TrustScore: 0.8},
	}
	jobs := []MatchableJob{
		// Job is >1km away → should produce no assignment
		{ID: "j1", Latitude: 29.0000, Longitude: 77.5000, RequiredSkills: []string{}, CreatedAt: now},
	}

	assignments := NearestNeighborMatch(workers, jobs, cfg, now)
	if len(assignments) != 0 {
		t.Errorf("expected no assignment (worker too far), got %d", len(assignments))
	}
}

func TestNearestNeighborMatch_NoDoubleBooking(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	// One worker, two nearby jobs — only 1 assignment allowed
	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6100, Longitude: 77.2000, IsOnline: true, TrustScore: 0.8},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6101, Longitude: 77.2001, RequiredSkills: []string{}, CreatedAt: now.Add(-1 * time.Minute)},
		{ID: "j2", Latitude: 28.6102, Longitude: 77.2002, RequiredSkills: []string{}, CreatedAt: now.Add(-2 * time.Minute)},
	}

	assignments := NearestNeighborMatch(workers, jobs, cfg, now)
	if len(assignments) > 1 {
		t.Errorf("expected at most 1 assignment from 1 worker, got %d", len(assignments))
	}
}

func TestNearestNeighborMatch_Speed(t *testing.T) {
	// Verify O(n) matcher completes quickly for large inputs
	cfg := DefaultConfig()
	now := time.Now()

	workers := make([]MatchableWorker, 200)
	for i := range workers {
		workers[i] = MatchableWorker{
			ID:         "w" + string(rune('A'+i%26)),
			Latitude:   28.60 + float64(i)*0.001,
			Longitude:  77.20 + float64(i)*0.001,
			IsOnline:   true,
			TrustScore: 0.75,
		}
	}
	jobs := make([]MatchableJob, 200)
	for i := range jobs {
		jobs[i] = MatchableJob{
			ID:        "j" + string(rune('A'+i%26)),
			Latitude:  28.60 + float64(i)*0.0011,
			Longitude: 77.20 + float64(i)*0.0011,
			CreatedAt: now.Add(-time.Duration(i) * time.Minute),
		}
	}

	start := time.Now()
	assignments := NearestNeighborMatch(workers, jobs, cfg, now)
	elapsed := time.Since(start)

	t.Logf("NearestNeighbor: 200×200 → %d assignments in %v", len(assignments), elapsed)
	if elapsed > 50*time.Millisecond {
		t.Errorf("NearestNeighbor 200×200 should complete in <50ms, took %v", elapsed)
	}
}
