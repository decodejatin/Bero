package matching

// =============================================================================
// Hungarian Algorithm — O(n³) Maximum Weight Bipartite Matching
//
// Converts max-weight to min-cost, handles non-square matrices via zero-padding.
// Stateless and safe for concurrent invocation.
//
// Reference: Kuhn-Munkres algorithm
// Constraint: Optimized for ≤ 50×50 matrices (< 10ms at this scale)
// =============================================================================

import "math"

// Assignment represents a row→col assignment with its weight.
type Assignment struct {
	Row    int
	Col    int
	Weight float64
}

// Solve finds the maximum weight bipartite matching.
//
// Input: weights[i][j] = score for assigning row i to col j (higher = better)
// Output: optimal assignments maximizing total weight
//
// Steps:
// 1. Pad to square matrix if needed
// 2. Convert max-weight → min-cost: cost[i][j] = maxVal - weights[i][j]
// 3. Run Hungarian on cost matrix
// 4. Return assignments with original weights (skip padded)
func Solve(weights [][]float64) []Assignment {
	rows := len(weights)
	if rows == 0 {
		return nil
	}
	cols := len(weights[0])
	if cols == 0 {
		return nil
	}

	// Determine square dimension
	n := rows
	if cols > n {
		n = cols
	}

	// Find max value for cost conversion
	maxVal := math.Inf(-1)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if weights[i][j] > maxVal {
				maxVal = weights[i][j]
			}
		}
	}
	if maxVal <= 0 {
		maxVal = 1.0
	}

	// Build square cost matrix: cost = maxVal - weight (padded cells = maxVal → zero weight)
	cost := make([][]float64, n)
	for i := 0; i < n; i++ {
		cost[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i < rows && j < cols {
				cost[i][j] = maxVal - weights[i][j]
			} else {
				cost[i][j] = maxVal // Padded: effectively zero weight
			}
		}
	}

	// Run Hungarian algorithm (Kuhn-Munkres)
	assignment := hungarian(cost, n)

	// Extract valid assignments (skip padded rows/cols)
	result := make([]Assignment, 0, rows)
	for i := 0; i < rows; i++ {
		j := assignment[i]
		if j < cols && j >= 0 {
			result = append(result, Assignment{
				Row:    i,
				Col:    j,
				Weight: weights[i][j],
			})
		}
	}

	return result
}

// hungarian implements the Kuhn-Munkres algorithm on an n×n cost matrix.
// Returns assignment[i] = column assigned to row i.
func hungarian(cost [][]float64, n int) []int {
	const INF = math.MaxFloat64

	// u[i], v[j] = potentials for rows and columns
	u := make([]float64, n+1)
	v := make([]float64, n+1)
	// p[j] = row assigned to column j (1-indexed, 0 = unassigned)
	p := make([]int, n+1)
	// way[j] = for column j, the predecessor column in the augmenting path
	way := make([]int, n+1)

	for i := 1; i <= n; i++ {
		// Start augmenting path from row i
		p[0] = i
		j0 := 0 // Virtual column

		minv := make([]float64, n+1)
		used := make([]bool, n+1)
		for j := 0; j <= n; j++ {
			minv[j] = INF
			used[j] = false
		}

		// Find augmenting path
		for {
			used[j0] = true
			i0 := p[j0]
			delta := INF
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
				break // Found free column, augmenting path complete
			}
		}

		// Trace back the augmenting path
		for j0 != 0 {
			p[j0] = p[way[j0]]
			j0 = way[j0]
		}
	}

	// Extract assignment: row i → column assignment[i]
	assignment := make([]int, n)
	for j := 1; j <= n; j++ {
		if p[j] > 0 {
			assignment[p[j]-1] = j - 1
		}
	}

	return assignment
}
