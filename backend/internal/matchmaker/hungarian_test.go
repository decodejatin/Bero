package matchmaker

import (
	"testing"
	"time"
)

func TestHungarianSolve_Square(t *testing.T) {
	// 3x3 classic assignment problem
	workers := []MatchableWorker{
		{ID: "w1"}, {ID: "w2"}, {ID: "w3"},
	}
	jobs := []MatchableJob{
		{ID: "j1"}, {ID: "j2"}, {ID: "j3"},
	}

	// Weight matrix designed so optimal is w1→j2, w2→j3, w3→j1
	weightMatrix := [][]float64{
		{0.5, 0.9, 0.3}, // w1: best on j2
		{0.2, 0.4, 0.8}, // w2: best on j3
		{0.7, 0.1, 0.2}, // w3: best on j1
	}
	// Optimal total = 0.9 + 0.8 + 0.7 = 2.4

	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, 0.0)

	if len(assignments) != 3 {
		t.Fatalf("expected 3 assignments, got %d", len(assignments))
	}

	totalWeight := 0.0
	for _, a := range assignments {
		totalWeight += a.Weight
	}

	if totalWeight < 2.39 || totalWeight > 2.41 {
		t.Errorf("expected total weight ~2.4, got %f", totalWeight)
	}

	t.Logf("Assignments: %+v, total: %f", assignments, totalWeight)
}

func TestHungarianSolve_MoreWorkersThanJobs(t *testing.T) {
	workers := []MatchableWorker{
		{ID: "w1"}, {ID: "w2"}, {ID: "w3"},
	}
	jobs := []MatchableJob{
		{ID: "j1"}, {ID: "j2"},
	}

	weightMatrix := [][]float64{
		{0.9, 0.1},
		{0.1, 0.9},
		{0.5, 0.5},
	}

	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, 0.0)

	// At most 2 assignments (limited by jobs)
	if len(assignments) > 2 {
		t.Fatalf("expected at most 2 assignments, got %d", len(assignments))
	}

	// Optimal: w1→j1 (0.9), w2→j2 (0.9) = 1.8
	totalWeight := 0.0
	for _, a := range assignments {
		totalWeight += a.Weight
	}

	if totalWeight < 1.79 || totalWeight > 1.81 {
		t.Errorf("expected total weight ~1.8, got %f", totalWeight)
	}

	t.Logf("Assignments: %+v, total: %f", assignments, totalWeight)
}

func TestHungarianSolve_MoreJobsThanWorkers(t *testing.T) {
	workers := []MatchableWorker{
		{ID: "w1"}, {ID: "w2"},
	}
	jobs := []MatchableJob{
		{ID: "j1"}, {ID: "j2"}, {ID: "j3"},
	}

	weightMatrix := [][]float64{
		{0.3, 0.9, 0.1},
		{0.1, 0.2, 0.8},
	}

	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, 0.0)

	// At most 2 assignments (limited by workers)
	if len(assignments) > 2 {
		t.Fatalf("expected at most 2 assignments, got %d", len(assignments))
	}

	// Optimal: w1→j2 (0.9), w2→j3 (0.8) = 1.7
	totalWeight := 0.0
	for _, a := range assignments {
		totalWeight += a.Weight
	}

	if totalWeight < 1.69 || totalWeight > 1.71 {
		t.Errorf("expected total weight ~1.7, got %f", totalWeight)
	}

	t.Logf("Assignments: %+v, total: %f", assignments, totalWeight)
}

func TestHungarianSolve_SinglePair(t *testing.T) {
	workers := []MatchableWorker{{ID: "w1"}}
	jobs := []MatchableJob{{ID: "j1"}}
	weightMatrix := [][]float64{{0.75}}

	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, 0.0)
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
	if assignments[0].WorkerID != "w1" || assignments[0].JobID != "j1" {
		t.Errorf("unexpected assignment: %+v", assignments[0])
	}
	if assignments[0].Weight != 0.75 {
		t.Errorf("expected weight 0.75, got %f", assignments[0].Weight)
	}
}

func TestHungarianSolve_EmptyInput(t *testing.T) {
	assignments := SolveMaxWeightMatching(nil, nil, nil, 0.0)
	if len(assignments) != 0 {
		t.Errorf("expected 0 assignments for empty input, got %d", len(assignments))
	}
}

func TestHungarianSolve_MinWeightThreshold(t *testing.T) {
	workers := []MatchableWorker{{ID: "w1"}, {ID: "w2"}}
	jobs := []MatchableJob{{ID: "j1"}, {ID: "j2"}}

	weightMatrix := [][]float64{
		{0.9, 0.05}, // w1-j2 below threshold
		{0.05, 0.9}, // w2-j1 below threshold
	}

	// With threshold 0.1, only diagonal assignments should survive
	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, 0.1)

	for _, a := range assignments {
		if a.Weight < 0.1 {
			t.Errorf("assignment below threshold: %+v", a)
		}
	}
}

func TestHungarianSolve_AllZeroWeights(t *testing.T) {
	workers := []MatchableWorker{{ID: "w1"}, {ID: "w2"}}
	jobs := []MatchableJob{{ID: "j1"}, {ID: "j2"}}

	weightMatrix := [][]float64{
		{0.0, 0.0},
		{0.0, 0.0},
	}

	// With minWeight=0, all zero weights are still valid
	assignments := SolveMaxWeightMatching(weightMatrix, workers, jobs, 0.0)
	// Any valid matching is acceptable
	t.Logf("Zero-weight assignments: %+v", assignments)
}

func TestEndToEnd_WeightAndHungarian(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
		{ID: "w3", Latitude: 28.6300, Longitude: 77.2200, Skills: []string{"PLUMBING", "ELECTRICAL"}, RatingAvg: 3.5},
	}

	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now.Add(-10 * time.Minute)},
	}

	matrix := BuildWeightMatrix(workers, jobs, cfg, now)
	assignments := SolveMaxWeightMatching(matrix, workers, jobs, cfg.MinWeightThreshold)

	// Should produce 2 assignments (limited by jobs)
	if len(assignments) > 2 {
		t.Fatalf("expected at most 2 assignments, got %d", len(assignments))
	}

	t.Logf("End-to-end assignments:")
	totalWeight := 0.0
	for _, a := range assignments {
		t.Logf("  Worker %s → Job %s (weight: %.4f)", a.WorkerID, a.JobID, a.Weight)
		totalWeight += a.Weight
	}
	t.Logf("  Total surplus S = %.4f", totalWeight)
}
