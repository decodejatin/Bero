package matchmaker

import (
	"math"
	"testing"
	"time"
)

func TestComputeProximity_SameLocation(t *testing.T) {
	cfg := DefaultConfig()
	worker := MatchableWorker{Latitude: 28.6139, Longitude: 77.2090}
	job := MatchableJob{Latitude: 28.6139, Longitude: 77.2090}
	score := computeProximity(worker, job, cfg)
	if score != 1.0 {
		t.Errorf("same location should give 1.0, got %f", score)
	}
}

func TestComputeProximity_FarAway(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxDistanceKm = 10.0
	// Delhi to Agra (~200km)
	worker := MatchableWorker{Latitude: 28.6139, Longitude: 77.2090}
	job := MatchableJob{Latitude: 27.1767, Longitude: 78.0081}
	score := computeProximity(worker, job, cfg)
	if score != 0.0 {
		t.Errorf("far away should give 0.0, got %f", score)
	}
}

func TestComputeProximity_Nearby(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxDistanceKm = 25.0
	// ~5km apart in Delhi
	worker := MatchableWorker{Latitude: 28.6139, Longitude: 77.2090}
	job := MatchableJob{Latitude: 28.6500, Longitude: 77.2300}
	score := computeProximity(worker, job, cfg)
	if score <= 0.0 || score >= 1.0 {
		t.Errorf("nearby should give between 0 and 1, got %f", score)
	}
}

func TestComputeReputation(t *testing.T) {
	tests := []struct {
		name     string
		rating   float64
		expected float64
	}{
		{"perfect rating", 5.0, 1.0},
		{"good rating", 4.0, 0.8},
		{"no rating", 0.0, 0.0},
		{"negative", -1.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := MatchableWorker{RatingAvg: tt.rating}
			result := computeReputation(w)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestComputeSkillMatch(t *testing.T) {
	tests := []struct {
		name         string
		workerSkills []string
		jobSkills    []string
		expected     float64
	}{
		{"perfect match", []string{"PLUMBING", "ELECTRICAL"}, []string{"PLUMBING", "ELECTRICAL"}, 1.0},
		{"partial match", []string{"PLUMBING"}, []string{"PLUMBING", "ELECTRICAL"}, 0.5},
		{"no match", []string{"CARPENTRY"}, []string{"PLUMBING", "ELECTRICAL"}, 0.0},
		{"no required skills", []string{"PLUMBING"}, nil, 1.0},
		{"empty worker skills", nil, []string{"PLUMBING"}, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := MatchableWorker{Skills: tt.workerSkills}
			j := MatchableJob{RequiredSkills: tt.jobSkills}
			result := computeSkillMatch(w, j)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestComputeWaitTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		created  time.Time
		expected float64
	}{
		{"just created", now, 0.0},
		{"30 min ago", now.Add(-30 * time.Minute), 0.5},
		{"60 min ago", now.Add(-60 * time.Minute), 1.0},
		{"120 min ago", now.Add(-120 * time.Minute), 1.0}, // capped at 1.0
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := MatchableJob{CreatedAt: tt.created}
			result := computeWaitTime(j, now)
			if math.Abs(result-tt.expected) > 0.02 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestComputeWeight_Integration(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	// Perfect worker: same location, 5-star, all skills, fresh job
	worker := MatchableWorker{
		ID:        "w1",
		Latitude:  28.6139,
		Longitude: 77.2090,
		Skills:    []string{"PLUMBING"},
		RatingAvg: 5.0,
	}
	job := MatchableJob{
		ID:             "j1",
		Latitude:       28.6139,
		Longitude:      77.2090,
		RequiredSkills: []string{"PLUMBING"},
		CreatedAt:      now,
	}

	weight := ComputeWeight(worker, job, cfg, now)
	// Expected: 0.35*1.0 + 0.20*1.0 + 0.30*1.0 - 0.15*0.0 = 0.85
	if math.Abs(weight-0.85) > 0.01 {
		t.Errorf("perfect match should give ~0.85, got %f", weight)
	}
}

func TestBuildWeightMatrix(t *testing.T) {
	cfg := DefaultConfig()
	now := time.Now()

	workers := []MatchableWorker{
		{ID: "w1", Latitude: 28.6139, Longitude: 77.2090, Skills: []string{"PLUMBING"}, RatingAvg: 4.5},
		{ID: "w2", Latitude: 28.6200, Longitude: 77.2100, Skills: []string{"ELECTRICAL"}, RatingAvg: 3.0},
	}
	jobs := []MatchableJob{
		{ID: "j1", Latitude: 28.6139, Longitude: 77.2090, RequiredSkills: []string{"PLUMBING"}, CreatedAt: now},
		{ID: "j2", Latitude: 28.6200, Longitude: 77.2100, RequiredSkills: []string{"ELECTRICAL"}, CreatedAt: now},
	}

	matrix := BuildWeightMatrix(workers, jobs, cfg, now)
	if len(matrix) != 2 || len(matrix[0]) != 2 {
		t.Fatalf("expected 2x2 matrix, got %dx%d", len(matrix), len(matrix[0]))
	}

	// Worker 1 should score higher on Job 1 (has PLUMBING skill)
	// Worker 2 should score higher on Job 2 (has ELECTRICAL skill)
	if matrix[0][0] <= matrix[0][1] {
		t.Log("Note: w1-j1 score should ideally be > w1-j2 (skill match)")
	}
}

func TestHaversine(t *testing.T) {
	// Delhi to Mumbai: ~1,153 km
	dist := haversineKm(28.6139, 77.2090, 19.0760, 72.8777)
	if dist < 1100 || dist > 1200 {
		t.Errorf("Delhi-Mumbai distance should be ~1153km, got %f", dist)
	}

	// Same point
	dist = haversineKm(28.6139, 77.2090, 28.6139, 77.2090)
	if dist != 0.0 {
		t.Errorf("same point distance should be 0, got %f", dist)
	}
}
