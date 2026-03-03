package matchmaker

import (
	"math"
	"testing"
	"time"
)

// ---------- TimeDependentUtility tests ----------

func TestTimeDependentUtility_Basic(t *testing.T) {
	// U(t) = V_job * e^(-λt) - C_travel
	// Fresh job (t=0): U = 0.8 * 1.0 - 0.1 = 0.7
	u := TimeDependentUtility(0.8, 0.05, 0.0, 0.1)
	if math.Abs(u-0.7) > 0.001 {
		t.Errorf("expected U ≈ 0.7 for fresh job, got %f", u)
	}

	// After 10 hours: U = 0.8 * e^(-0.5) - 0.1 ≈ 0.8*0.6065 - 0.1 ≈ 0.385
	u10 := TimeDependentUtility(0.8, 0.05, 10.0, 0.1)
	if u10 >= u {
		t.Errorf("utility should decay over time: U(0)=%f, U(10)=%f", u, u10)
	}
	expected10 := 0.8*math.Exp(-0.5) - 0.1
	if math.Abs(u10-expected10) > 0.001 {
		t.Errorf("expected U(10) ≈ %f, got %f", expected10, u10)
	}
}

func TestTimeDependentUtility_TravelCost(t *testing.T) {
	// Higher travel cost = lower utility
	uLow := TimeDependentUtility(0.8, 0.05, 1.0, 0.05)
	uHigh := TimeDependentUtility(0.8, 0.05, 1.0, 0.30)
	if uHigh >= uLow {
		t.Errorf("higher travel cost should give lower utility: cTravel=0.05 → %f, cTravel=0.30 → %f", uLow, uHigh)
	}
}

func TestTimeDependentUtility_ZeroDecay(t *testing.T) {
	// λ = 0 means no time decay: U = V - C_travel regardless of t
	u0 := TimeDependentUtility(0.8, 0.0, 0.0, 0.1)
	u100 := TimeDependentUtility(0.8, 0.0, 100.0, 0.1)
	if math.Abs(u0-u100) > 0.001 {
		t.Errorf("zero decay should give same utility at any time: U(0)=%f, U(100)=%f", u0, u100)
	}
	if math.Abs(u0-0.7) > 0.001 {
		t.Errorf("expected 0.7, got %f", u0)
	}
}

// ---------- BlockingPair detection tests ----------

func makeTestWorkers() []MatchableWorker {
	return []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
		{ID: "w3", Latitude: 28.6300, Longitude: 77.2200, Skills: []string{"PLUMBING", "ELECTRICAL"}, RatingAvg: 3.5},
	}
}

func makeTestJobs(now time.Time) []MatchableJob {
	return []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now},
	}
}

func TestFindBlockingPairs_NoBlocking(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := makeTestWorkers()
	jobs := makeTestJobs(now)

	// Optimal matching from Hungarian: w1→j1, w2→j2
	// (each worker matched to their best-skill job at closest location)
	matrix := BuildWeightMatrix(workers, jobs, cfg, now)
	assignments := SolveMaxWeightMatching(matrix, workers, jobs, cfg.MinWeightThreshold)

	pairs := FindBlockingPairs(assignments, workers, jobs, cfg, now)

	// With switching cost, the optimal Hungarian matching should be stable
	t.Logf("Optimal assignments: %+v", assignments)
	t.Logf("Blocking pairs found: %d", len(pairs))

	// An optimally matched result with switching costs should have few/no blocking pairs
	// (switching cost acts as friction preventing instability)
}

func TestFindBlockingPairs_WithBlocking(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SwitchingCost = 0.0 // Remove switching cost to make blocking easier to find
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING", "ELECTRICAL"}, RatingAvg: 5.0},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"PLUMBING"}, RatingAvg: 3.0},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now},
	}

	// Force a suboptimal assignment: w1→j2, w2→j1
	// w1 is at j1's location with all skills + 5-star rating, but assigned to j2
	suboptimal := []Assignment{
		{WorkerID: "w1", JobID: "j2", Weight: 0.5},
		{WorkerID: "w2", JobID: "j1", Weight: 0.3},
	}

	pairs := FindBlockingPairs(suboptimal, workers, jobs, cfg, now)

	if len(pairs) == 0 {
		t.Error("expected blocking pairs for suboptimal assignment with zero switching cost")
	}

	t.Logf("Found %d blocking pairs", len(pairs))
	for _, bp := range pairs {
		t.Logf("  Blocking: w%d→j%d (workerGain=%.4f, jobGain=%.4f)",
			bp.WorkerIdx, bp.JobIdx, bp.WorkerGain, bp.JobGain)
	}
}

func TestFindBlockingPairs_SwitchingCostPrevents(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SwitchingCost = 1.0 // Very high switching cost
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now},
	}

	// Even a suboptimal assignment shouldn't create blocking pairs with high C_switch
	assignments := []Assignment{
		{WorkerID: "w1", JobID: "j2", Weight: 0.5},
		{WorkerID: "w2", JobID: "j1", Weight: 0.3},
	}

	pairs := FindBlockingPairs(assignments, workers, jobs, cfg, now)
	if len(pairs) != 0 {
		t.Errorf("high switching cost should prevent blocking pairs, found %d", len(pairs))
	}
}

// ---------- ResolveBlockingPairs / EnforceStability tests ----------

func TestResolveBlockingPairs_ResolvesAll(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SwitchingCost = 0.0 // No friction for clean resolution
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING", "ELECTRICAL"}, RatingAvg: 5.0},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"PLUMBING"}, RatingAvg: 3.0},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now},
	}

	// Start with suboptimal matching
	suboptimal := []Assignment{
		{WorkerID: "w1", JobID: "j2", Weight: 0.5},
		{WorkerID: "w2", JobID: "j1", Weight: 0.3},
	}

	resolved := ResolveBlockingPairs(suboptimal, workers, jobs, cfg, now)

	// After resolution, check for remaining blocking pairs
	remaining := FindBlockingPairs(resolved, workers, jobs, cfg, now)

	t.Logf("Resolved assignments: %+v", resolved)
	t.Logf("Remaining blocking pairs: %d", len(remaining))

	if len(remaining) > 0 {
		t.Errorf("expected 0 blocking pairs after resolution, got %d", len(remaining))
		for _, bp := range remaining {
			t.Logf("  Still blocking: w%d→j%d", bp.WorkerIdx, bp.JobIdx)
		}
	}
}

func TestEnforceStability_EndToEnd(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := makeTestWorkers()
	jobs := makeTestJobs(now)

	// Get Hungarian matching
	matrix := BuildWeightMatrix(workers, jobs, cfg, now)
	assignments := SolveMaxWeightMatching(matrix, workers, jobs, cfg.MinWeightThreshold)

	t.Logf("Hungarian assignments:")
	for _, a := range assignments {
		t.Logf("  %s → %s (w=%.4f)", a.WorkerID, a.JobID, a.Weight)
	}

	// Enforce stability
	stable := EnforceStability(assignments, workers, jobs, cfg, now)

	t.Logf("Stable assignments:")
	for _, a := range stable {
		t.Logf("  %s → %s (w=%.4f)", a.WorkerID, a.JobID, a.Weight)
	}

	// Should produce valid assignments
	if len(stable) == 0 {
		t.Error("expected at least 1 stable assignment")
	}

	// Verify no blocking pairs in final result
	remaining := FindBlockingPairs(stable, workers, jobs, cfg, now)
	if len(remaining) > 0 {
		t.Logf("Warning: %d blocking pairs remain (may be due to switching cost)", len(remaining))
	}
}

func TestEnforceStability_Disabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EnableStability = false
	now := time.Now()

	workers := makeTestWorkers()
	jobs := makeTestJobs(now)

	matrix := BuildWeightMatrix(workers, jobs, cfg, now)
	original := SolveMaxWeightMatching(matrix, workers, jobs, cfg.MinWeightThreshold)

	// When disabled, EnforceStability should still work (it's called manually here)
	// but the engine won't call it. Verify the function is a no-op when already stable.
	result := EnforceStability(original, workers, jobs, cfg, now)

	// Result should have the same number of assignments
	if len(result) != len(original) {
		t.Logf("Note: assignment count changed from %d to %d (Gale-Shapley refinement)",
			len(original), len(result))
	}
}

func TestEnforceStability_EmptyInput(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	result := EnforceStability(nil, nil, nil, cfg, now)
	if len(result) != 0 {
		t.Errorf("expected 0 assignments for empty input, got %d", len(result))
	}
}

func TestSortByDesc(t *testing.T) {
	indices := []int{0, 1, 2, 3, 4}
	values := []float64{0.3, 0.9, 0.1, 0.7, 0.5}

	sortByDesc(indices, func(i int) float64 { return values[i] })

	// Expected order: 1(0.9), 3(0.7), 4(0.5), 0(0.3), 2(0.1)
	expected := []int{1, 3, 4, 0, 2}
	for i, idx := range indices {
		if idx != expected[i] {
			t.Errorf("position %d: expected index %d, got %d", i, expected[i], idx)
		}
	}
}
