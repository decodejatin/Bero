package matchmaker

import "math"

// SolveMaxWeightMatching finds the maximum weight matching in a bipartite graph
// using the Hungarian Algorithm (Kuhn-Munkres).
//
// weightMatrix[i][j] is the weight of edge (worker_i, job_j).
// The matrix may be rectangular (|workers| ≠ |jobs|).
//
// Returns a list of assignments that maximizes total surplus:
//
//	S = Σ x_ij · w_ij
//
// Subject to:
//
//	Σ_j x_ij ≤ 1  ∀ i ∈ W  (each worker assigned at most 1 job)
//	Σ_i x_ij ≤ 1  ∀ j ∈ J  (each job assigned at most 1 worker)
//
// Assignments with weight ≤ minWeight are filtered out.
func SolveMaxWeightMatching(weightMatrix [][]float64, workers []MatchableWorker, jobs []MatchableJob, minWeight float64) []Assignment {
	nWorkers := len(workers)
	nJobs := len(jobs)

	if nWorkers == 0 || nJobs == 0 {
		return nil
	}

	// Pad to square matrix (Hungarian requires square)
	n := nWorkers
	if nJobs > n {
		n = nJobs
	}

	// Build cost matrix: negate weights to convert max-weight to min-cost.
	// Find the maximum weight to shift all values positive (Hungarian needs non-negative costs).
	maxWeight := math.Inf(-1)
	for i := 0; i < nWorkers; i++ {
		for j := 0; j < nJobs; j++ {
			if weightMatrix[i][j] > maxWeight {
				maxWeight = weightMatrix[i][j]
			}
		}
	}
	if maxWeight < 0 {
		maxWeight = 0
	}

	// Create padded square cost matrix: cost = maxWeight - weight
	// Dummy rows/cols get cost = maxWeight (i.e., weight = 0)
	cost := make([][]float64, n)
	for i := 0; i < n; i++ {
		cost[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i < nWorkers && j < nJobs {
				cost[i][j] = maxWeight - weightMatrix[i][j]
			} else {
				cost[i][j] = maxWeight // dummy edge
			}
		}
	}

	// Run Hungarian algorithm
	assignment := hungarianSolve(cost, n)

	// Extract valid assignments (non-dummy, above threshold)
	var result []Assignment
	for i := 0; i < nWorkers; i++ {
		j := assignment[i]
		if j < nJobs {
			w := weightMatrix[i][j]
			if w >= minWeight {
				result = append(result, Assignment{
					WorkerID: workers[i].ID,
					JobID:    jobs[j].ID,
					Weight:   w,
				})
			}
		}
	}

	return result
}

// hungarianSolve implements the Hungarian algorithm for minimum cost assignment.
// cost is an n×n square matrix. Returns assignment[i] = column assigned to row i.
func hungarianSolve(cost [][]float64, n int) []int {
	const inf = math.MaxFloat64

	// u[i] and v[j] are potentials for rows and columns (1-indexed internally)
	u := make([]float64, n+1)
	v := make([]float64, n+1)
	// p[j] = row assigned to column j
	p := make([]int, n+1)
	// way[j] = column that led to column j in the augmenting path
	way := make([]int, n+1)

	for i := 1; i <= n; i++ {
		p[0] = i
		j0 := 0 // virtual zero column
		minv := make([]float64, n+1)
		used := make([]bool, n+1)
		for j := 0; j <= n; j++ {
			minv[j] = inf
			used[j] = false
		}

		// Find augmenting path
		for {
			used[j0] = true
			i0 := p[j0]
			delta := inf
			j1 := -1

			for j := 1; j <= n; j++ {
				if !used[j] {
					cur := cost[i0-1][j-1] - u[i0] - v[j]
					if cur < minv[j] {
						minv[j] = cur
						way[j] = j0
					}
					if minv[j] < delta {
						delta = minv[j]
						j1 = j
					}
				}
			}

			// Update potentials
			for j := 0; j <= n; j++ {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}

			j0 = j1
			if p[j0] == 0 {
				break
			}
		}

		// Reconstruct path
		for j0 != 0 {
			p[j0] = p[way[j0]]
			j0 = way[j0]
		}
	}

	// Build result: assignment[row] = col (0-indexed)
	result := make([]int, n)
	for j := 1; j <= n; j++ {
		if p[j] != 0 {
			result[p[j]-1] = j - 1
		}
	}

	return result
}
