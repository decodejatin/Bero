package matchmaker

import (
	"math"
	"time"
)

// ComputeWeight calculates the composite weight w_ij for a worker-job pair:
//
//	w_ij = α1·Proximity_ij + α2·Reputation_i + α3·SkillMatch_ij − α4·WaitTime_j
func ComputeWeight(worker MatchableWorker, job MatchableJob, cfg MatchConfig, now time.Time) float64 {
	proximity := computeProximity(worker, job, cfg)
	// §6.1: Use Bayesian TrustScore instead of raw RatingAvg.
	// TrustScore is the Wilson Score CI lower bound on Beta(α₀+S, β₀+F),
	// which penalizes uncertainty and rewards review volume over lucky averages.
	reputation := computeReputation(worker)
	skillMatch := computeSkillMatch(worker, job)
	waitPenalty := computeWaitTime(job, now)

	weight := cfg.AlphaProximity*proximity +
		cfg.AlphaReputation*reputation +
		cfg.AlphaSkillMatch*skillMatch -
		cfg.AlphaWaitTime*waitPenalty

	return weight
}

// BuildWeightMatrix constructs an |W|×|J| weight matrix for all worker-job pairs.
func BuildWeightMatrix(workers []MatchableWorker, jobs []MatchableJob, cfg MatchConfig, now time.Time) [][]float64 {
	n := len(workers)
	m := len(jobs)
	matrix := make([][]float64, n)
	for i := 0; i < n; i++ {
		matrix[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			matrix[i][j] = ComputeWeight(workers[i], jobs[j], cfg, now)
		}
	}
	return matrix
}

// computeProximity returns a [0,1] score where 1 = same location, 0 = beyond max distance.
// Uses the Haversine formula for great-circle distance.
func computeProximity(worker MatchableWorker, job MatchableJob, cfg MatchConfig) float64 {
	dist := haversineKm(worker.Latitude, worker.Longitude, job.Latitude, job.Longitude)
	if dist >= cfg.MaxDistanceKm {
		return 0.0
	}
	// Linear decay: closer = higher score
	return 1.0 - (dist / cfg.MaxDistanceKm)
}

// computeReputation returns a [0,1] reputation score for the worker.
// §6.1: Prefers the Bayesian TrustScore (Wilson Score CI lower bound) when
// available (TrustScore > 0), which statistically penalizes low-review-volume
// workers. Falls back to raw RatingAvg/5.0 for workers without a TrustScore.
func computeReputation(worker MatchableWorker) float64 {
	if worker.TrustScore > 0 {
		// TrustScore is already normalized to [0,1]
		return math.Min(worker.TrustScore, 1.0)
	}
	// Fallback: normalize raw 5-star average
	if worker.RatingAvg <= 0 {
		return 0.5 // neutral prior for unrated workers
	}
	rating := worker.RatingAvg / 5.0
	if rating > 1.0 {
		rating = 1.0
	}
	return rating
}

// computeSkillMatch returns the fraction of required skills the worker possesses.
// Returns 1.0 if job has no required skills (any worker fits).
func computeSkillMatch(worker MatchableWorker, job MatchableJob) float64 {
	if len(job.RequiredSkills) == 0 {
		return 1.0
	}

	workerSkillSet := make(map[string]struct{}, len(worker.Skills))
	for _, s := range worker.Skills {
		workerSkillSet[s] = struct{}{}
	}

	matched := 0
	for _, rs := range job.RequiredSkills {
		if _, ok := workerSkillSet[rs]; ok {
			matched++
		}
	}

	return float64(matched) / float64(len(job.RequiredSkills))
}

// computeWaitTime returns a [0,1] value representing how long the job has waited.
// 0 = just created, 1 = waited ≥ 60 minutes.
// Higher values mean the job has waited longer (applied as negative penalty in weight formula).
func computeWaitTime(job MatchableJob, now time.Time) float64 {
	elapsed := now.Sub(job.CreatedAt).Minutes()
	if elapsed <= 0 {
		return 0.0
	}
	penalty := elapsed / 60.0
	if penalty > 1.0 {
		penalty = 1.0
	}
	return penalty
}

// haversineKm calculates the great-circle distance in kilometers between two lat/lng points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}
