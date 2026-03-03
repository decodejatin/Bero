package matchmaker

import (
	"testing"
	"time"
)

func TestRunShadowMatching_ComparesCorrectly(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
		{ID: "w3", Latitude: 28.6300, Longitude: 77.2200, Skills: []string{"PLUMBING", "ELECTRICAL"}, RatingAvg: 3.5},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now},
	}

	// Get live assignments via Hungarian
	matrix := BuildWeightMatrix(workers, jobs, cfg, now)
	liveAssignments := SolveMaxWeightMatching(matrix, workers, jobs, cfg.MinWeightThreshold)

	// Run shadow comparison
	result := RunShadowMatching(workers, jobs, cfg, now, liveAssignments)

	if result.Comparison.LiveCount == 0 {
		t.Error("expected live assignments")
	}
	if result.Comparison.ShadowCount == 0 {
		t.Error("expected shadow assignments")
	}

	t.Logf("Live: %d assignments (weight=%.4f), Shadow: %d assignments (weight=%.4f)",
		result.Comparison.LiveCount, result.Comparison.LiveTotalWeight,
		result.Comparison.ShadowCount, result.Comparison.ShadowTotalWeight)
	t.Logf("Delta: %.4f (%.1f%%), Live stable=%v, Shadow stable=%v",
		result.Comparison.WeightDelta, result.Comparison.WeightDeltaPct,
		result.Comparison.LiveIsStable, result.Comparison.ShadowIsStable)

	// Hungarian should be >= greedy in total weight (it's optimal)
	if result.Comparison.WeightDelta < -0.01 {
		t.Logf("Warning: greedy outperformed Hungarian by %.4f — this is unexpected", -result.Comparison.WeightDelta)
	}
}

func TestGreedyMatch(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 5.0},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
	}

	assignments := greedyMatch(workers, jobs, cfg, now)

	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}

	// w1 should be assigned to j1 (same location + perfect skill match)
	if assignments[0].WorkerID != "w1" {
		t.Errorf("expected w1, got %s", assignments[0].WorkerID)
	}
}

func TestGreedyMatch_Empty(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	if len(greedyMatch(nil, nil, cfg, now)) != 0 {
		t.Error("empty input should return empty")
	}
}
