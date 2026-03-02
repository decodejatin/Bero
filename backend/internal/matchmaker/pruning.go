package matchmaker

import (
	"sort"
	"time"

	"github.com/decodejatin/bero-backend/internal/matchmaker/h3"
)

// SpatialIndex is an in-memory index mapping H3 cells to worker indices.
// It provides O(1) spatial lookups for finding nearby workers.
type SpatialIndex struct {
	// cellToWorkers maps H3 cell index → list of worker indices in the workers slice...
	cellToWorkers map[h3.Cell][]int
	resolution    int
}

// NewSpatialIndex builds a spatial index from a list of workers.
func NewSpatialIndex(workers []MatchableWorker, resolution int) *SpatialIndex {
	idx := &SpatialIndex{
		cellToWorkers: make(map[h3.Cell][]int),
		resolution:    resolution,
	}

	for i, w := range workers {
		cell := w.H3Index
		if cell == "" {
			// Compute H3 index from lat/lng if not pre-computed
			cell = string(h3.LatLngToCell(w.Latitude, w.Longitude, resolution))
		}
		idx.cellToWorkers[h3.Cell(cell)] = append(idx.cellToWorkers[h3.Cell(cell)], i)
	}

	return idx
}

// FindCandidateIndices returns worker indices near the given job using H3 k-ring expansion.
// It starts from ring 1 and expands until at least k candidates are found or maxRings is reached.
func (idx *SpatialIndex) FindCandidateIndices(job MatchableJob, k int, maxRings int) []int {
	jobCell := h3.Cell(job.H3Index)
	if jobCell == "" {
		jobCell = h3.LatLngToCell(job.Latitude, job.Longitude, idx.resolution)
	}

	seen := make(map[int]struct{})
	var candidates []int

	for ring := 0; ring <= maxRings; ring++ {
		cells := h3.KRing(jobCell, ring)
		for _, cell := range cells {
			if workerIndices, ok := idx.cellToWorkers[cell]; ok {
				for _, wi := range workerIndices {
					if _, exists := seen[wi]; !exists {
						seen[wi] = struct{}{}
						candidates = append(candidates, wi)
					}
				}
			}
		}

		// Stop expanding once we have enough candidates
		if len(candidates) >= k {
			break
		}
	}

	return candidates
}

// WorkerCount returns the number of workers in a given H3 cell.
func (idx *SpatialIndex) WorkerCount(cell h3.Cell) int {
	return len(idx.cellToWorkers[cell])
}

// DensityMap returns a map of H3 cell → worker count (supply density).
func (idx *SpatialIndex) DensityMap() map[string]int {
	density := make(map[string]int, len(idx.cellToWorkers))
	for cell, workers := range idx.cellToWorkers {
		density[string(cell)] = len(workers)
	}
	return density
}

// PruneAndMatch performs dynamic candidate pruning followed by Hungarian matching.
//
// Instead of building a full W×J weight matrix (O(W×J)), for each job we:
// 1. Use the spatial index to find the k-nearest candidate workers (via H3 k-ring)
// 2. Build a reduced weight matrix with only the candidates
// 3. Run Hungarian algorithm on the reduced graph
//
// This reduces complexity from O(n³) to O(k³) where k << n.
func PruneAndMatch(workers []MatchableWorker, jobs []MatchableJob, cfg MatchConfig, now time.Time) []Assignment {
	if len(workers) == 0 || len(jobs) == 0 {
		return nil
	}

	// Build spatial index
	spatialIdx := NewSpatialIndex(workers, cfg.H3Resolution)

	// Determine k-ring expansion radius
	maxRings := h3.KRingRadiusForDistance(cfg.H3Resolution, cfg.MaxDistanceKm)
	k := cfg.KNearestNeighbors

	// For each job, find candidate workers
	// Use a greedy approach: assign best available worker per job
	// while respecting the constraint that each worker is assigned at most once
	type jobCandidates struct {
		jobIdx     int
		candidates []int // indices into workers slice
	}

	jobCands := make([]jobCandidates, len(jobs))
	for j := range jobs {
		candIndices := spatialIdx.FindCandidateIndices(jobs[j], k, maxRings)
		jobCands[j] = jobCandidates{
			jobIdx:     j,
			candidates: candIndices,
		}
	}

	// Collect all unique candidate workers across all jobs
	candidateSet := make(map[int]struct{})
	for _, jc := range jobCands {
		for _, wi := range jc.candidates {
			candidateSet[wi] = struct{}{}
		}
	}

	if len(candidateSet) == 0 {
		return nil
	}

	// Build reduced worker list and mapping
	candidateWorkers := make([]MatchableWorker, 0, len(candidateSet))
	originalIdx := make([]int, 0, len(candidateSet))      // reduced index → original index
	reducedIdxMap := make(map[int]int, len(candidateSet)) // original index → reduced index

	for oi := range candidateSet {
		reducedIdxMap[oi] = len(candidateWorkers)
		originalIdx = append(originalIdx, oi)
		candidateWorkers = append(candidateWorkers, workers[oi])
	}

	// Sort jobs by number of candidates (ascending) — jobs with fewer candidates get priority
	sortedJobs := make([]int, len(jobs))
	for i := range sortedJobs {
		sortedJobs[i] = i
	}
	sort.Slice(sortedJobs, func(a, b int) bool {
		return len(jobCands[sortedJobs[a]].candidates) < len(jobCands[sortedJobs[b]].candidates)
	})

	// Build reduced weight matrix (candidateWorkers × jobs)
	weightMatrix := BuildWeightMatrix(candidateWorkers, jobs, cfg, now)

	// Run Hungarian algorithm on the reduced graph
	assignments := SolveMaxWeightMatching(weightMatrix, candidateWorkers, jobs, cfg.MinWeightThreshold)

	// Map back to original worker IDs (already set by SolveMaxWeightMatching using worker.ID)
	return assignments
}

// ComputeSupplyDensity analyzes the spatial distribution of workers and returns
// density information per H3 cell.
type SupplyDensityInfo struct {
	Cell        string  `json:"cell"`
	WorkerCount int     `json:"worker_count"`
	CenterLat   float64 `json:"center_lat"`
	CenterLng   float64 `json:"center_lng"`
}

// GetSupplyDensity returns the supply density for all cells containing workers.
func GetSupplyDensity(workers []MatchableWorker, resolution int) []SupplyDensityInfo {
	spatialIdx := NewSpatialIndex(workers, resolution)
	density := spatialIdx.DensityMap()

	result := make([]SupplyDensityInfo, 0, len(density))
	for cellStr, count := range density {
		cell := h3.Cell(cellStr)
		center, err := h3.CellToLatLng(cell)
		if err != nil {
			continue
		}
		result = append(result, SupplyDensityInfo{
			Cell:        cellStr,
			WorkerCount: count,
			CenterLat:   center.Lat,
			CenterLng:   center.Lng,
		})
	}

	// Sort by worker count descending (hotspots first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].WorkerCount > result[j].WorkerCount
	})

	return result
}
