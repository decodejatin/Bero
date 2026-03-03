package matchmaker

import (
	"log"
	"time"
)

// ShadowComparison holds the comparison metrics between live and shadow matching.
type ShadowComparison struct {
	LiveTotalWeight   float64 `json:"live_total_weight"`
	ShadowTotalWeight float64 `json:"shadow_total_weight"`
	LiveCount         int     `json:"live_count"`
	ShadowCount       int     `json:"shadow_count"`
	WeightDelta       float64 `json:"weight_delta"`     // live - shadow
	WeightDeltaPct    float64 `json:"weight_delta_pct"` // percentage improvement
	LiveIsStable      bool    `json:"live_is_stable"`   // has no blocking pairs
	ShadowIsStable    bool    `json:"shadow_is_stable"`
}

// ShadowResult holds both live and shadow matching results for comparison.
type ShadowResult struct {
	LiveAssignments   []Assignment     `json:"live_assignments"`
	ShadowAssignments []Assignment     `json:"shadow_assignments"`
	Comparison        ShadowComparison `json:"comparison"`
}

// RunShadowMatching runs the main matching algorithm and a shadow (greedy) algorithm
// in parallel, then compares results. Shadow results are NEVER used for live assignments,
// they are logged for offline analysis and A/B testing.
func RunShadowMatching(
	workers []MatchableWorker,
	jobs []MatchableJob,
	cfg MatchConfig,
	now time.Time,
	liveAssignments []Assignment,
) ShadowResult {
	// Shadow strategy: simple greedy matching (fast, no Hungarian overhead)
	shadowAssignments := greedyMatch(workers, jobs, cfg, now)

	// Compare
	liveTotalWeight := totalWeight(liveAssignments)
	shadowTotalWeight := totalWeight(shadowAssignments)

	delta := liveTotalWeight - shadowTotalWeight
	deltaPct := 0.0
	if shadowTotalWeight > 0 {
		deltaPct = (delta / shadowTotalWeight) * 100.0
	}

	// Check stability of both
	liveBlocking := FindBlockingPairs(liveAssignments, workers, jobs, cfg, now)
	shadowBlocking := FindBlockingPairs(shadowAssignments, workers, jobs, cfg, now)

	comparison := ShadowComparison{
		LiveTotalWeight:   liveTotalWeight,
		ShadowTotalWeight: shadowTotalWeight,
		LiveCount:         len(liveAssignments),
		ShadowCount:       len(shadowAssignments),
		WeightDelta:       delta,
		WeightDeltaPct:    deltaPct,
		LiveIsStable:      len(liveBlocking) == 0,
		ShadowIsStable:    len(shadowBlocking) == 0,
	}

	log.Printf("[shadow] Live surplus=%.4f (%d assignments), Shadow surplus=%.4f (%d assignments), Δ=%.4f (%.1f%%)",
		liveTotalWeight, len(liveAssignments),
		shadowTotalWeight, len(shadowAssignments),
		delta, deltaPct)

	return ShadowResult{
		LiveAssignments:   liveAssignments,
		ShadowAssignments: shadowAssignments,
		Comparison:        comparison,
	}
}

// greedyMatch performs a simple greedy matching: for each job (in order), assign
// the best available worker. O(W×J) — much faster than Hungarian but suboptimal.
func greedyMatch(
	workers []MatchableWorker,
	jobs []MatchableJob,
	cfg MatchConfig,
	now time.Time,
) []Assignment {
	assigned := make(map[int]bool) // worker index → already assigned
	var result []Assignment

	for _, job := range jobs {
		bestIdx := -1
		bestWeight := cfg.MinWeightThreshold

		for wi, worker := range workers {
			if assigned[wi] {
				continue
			}
			w := ComputeWeight(worker, job, cfg, now)
			if w > bestWeight {
				bestWeight = w
				bestIdx = wi
			}
		}

		if bestIdx >= 0 {
			assigned[bestIdx] = true
			result = append(result, Assignment{
				WorkerID: workers[bestIdx].ID,
				JobID:    job.ID,
				Weight:   bestWeight,
			})
		}
	}

	return result
}

func totalWeight(assignments []Assignment) float64 {
	total := 0.0
	for _, a := range assignments {
		total += a.Weight
	}
	return total
}
