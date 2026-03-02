package matchmaker

import (
	"testing"
	"time"

	"github.com/decodejatin/bero-backend/internal/matchmaker/h3"
)

func TestSpatialIndex_Build(t *testing.T) {
	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
		{ID: "w3", Latitude: 19.0760, Longitude: 72.8777, Skills: []string{"PLUMBING"}, RatingAvg: 3.5}, // Mumbai — far away
	}

	idx := NewSpatialIndex(workers, h3.DefaultResolution)
	density := idx.DensityMap()

	t.Logf("Spatial index has %d cells", len(density))
	for cell, count := range density {
		t.Logf("  Cell %s: %d workers", cell, count)
	}

	if len(density) < 2 {
		t.Error("Delhi and Mumbai workers should be in different cells")
	}
}

func TestSpatialIndex_FindCandidates_NearbyWorkers(t *testing.T) {
	// Workers clustered in Delhi
	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6150, Longitude: 77.2095, Skills: []string{"ELECTRICAL"}, RatingAvg: 4.0},
		{ID: "w3", Latitude: 28.6170, Longitude: 77.2080, Skills: []string{"PLUMBING"}, RatingAvg: 3.5},
		{ID: "w4", Latitude: 19.0760, Longitude: 72.8777, Skills: []string{"PLUMBING"}, RatingAvg: 5.0}, // Mumbai
	}

	idx := NewSpatialIndex(workers, h3.DefaultResolution)

	// Job in Delhi
	job := MatchableJob{
		ID:        "j1",
		Latitude:  28.6145,
		Longitude: 77.2092,
	}

	candidates := idx.FindCandidateIndices(job, 3, 5)

	t.Logf("Found %d candidates for Delhi job", len(candidates))
	for _, ci := range candidates {
		t.Logf("  Candidate: %s (%.4f, %.4f)", workers[ci].ID, workers[ci].Latitude, workers[ci].Longitude)
	}

	// Mumbai worker should NOT be in candidates (too far)
	for _, ci := range candidates {
		if workers[ci].ID == "w4" {
			t.Error("Mumbai worker should not be a candidate for a Delhi job")
		}
	}
}

func TestSpatialIndex_FindCandidates_EmptyArea(t *testing.T) {
	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090},
	}

	idx := NewSpatialIndex(workers, h3.DefaultResolution)

	// Job in a completely different area (Mumbai)
	job := MatchableJob{
		ID:        "j1",
		Latitude:  19.0760,
		Longitude: 72.8777,
	}

	candidates := idx.FindCandidateIndices(job, 5, 3) // small maxRings

	t.Logf("Found %d candidates for distant job (expected 0)", len(candidates))
	// With small maxRings, should find no candidates in a distant area
}

func TestPruneAndMatch_Basic(t *testing.T) {
	cfg := DefaultConfig()
	cfg.H3Resolution = h3.DefaultResolution
	cfg.KNearestNeighbors = 5
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

	assignments := PruneAndMatch(workers, jobs, cfg, now)

	t.Logf("Pruned matching produced %d assignments:", len(assignments))
	for _, a := range assignments {
		t.Logf("  Worker %s → Job %s (weight: %.4f)", a.WorkerID, a.JobID, a.Weight)
	}

	if len(assignments) == 0 {
		t.Error("expected at least 1 assignment")
	}
}

func TestPruneAndMatch_EmptyInput(t *testing.T) {
	cfg := DefaultConfig()
	cfg.H3Resolution = h3.DefaultResolution
	cfg.KNearestNeighbors = 5
	now := time.Now()

	assignments := PruneAndMatch(nil, nil, cfg, now)
	if len(assignments) != 0 {
		t.Errorf("expected 0 assignments for empty input, got %d", len(assignments))
	}
}

func TestPruneAndMatch_FarAwayWorkersExcluded(t *testing.T) {
	cfg := DefaultConfig()
	cfg.H3Resolution = h3.DefaultResolution
	cfg.KNearestNeighbors = 3
	cfg.MaxDistanceKm = 5.0 // Only 5km radius
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w-delhi", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w-mumbai", Latitude: 19.0760, Longitude: 72.8777, Skills: []string{"PLUMBING"}, RatingAvg: 5.0},
	}

	jobs := []MatchableJob{
		{ID: "j-delhi", Latitude: 28.6140, Longitude: 77.2091, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
	}

	assignments := PruneAndMatch(workers, jobs, cfg, now)

	t.Logf("Assignments: %+v", assignments)

	// Mumbai worker should not be assigned to Delhi job
	for _, a := range assignments {
		if a.WorkerID == "w-mumbai" {
			t.Error("Mumbai worker should not be assigned to a Delhi job with 5km limit")
		}
	}
}

func TestGetSupplyDensity(t *testing.T) {
	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090},
		{ID: "w2", Latitude: 28.6140, Longitude: 77.2091},
		{ID: "w3", Latitude: 28.6200, Longitude: 77.2100},
		{ID: "w4", Latitude: 19.0760, Longitude: 72.8777},
	}

	density := GetSupplyDensity(workers, h3.DefaultResolution)

	t.Logf("Supply density (%d cells):", len(density))
	for _, d := range density {
		t.Logf("  Cell %s: %d workers at (%.4f, %.4f)", d.Cell, d.WorkerCount, d.CenterLat, d.CenterLng)
	}

	if len(density) < 2 {
		t.Error("expected at least 2 cells (Delhi cluster + Mumbai)")
	}

	// First entry should have the highest count
	if len(density) > 1 && density[0].WorkerCount < density[1].WorkerCount {
		t.Error("density should be sorted by worker count descending")
	}
}
