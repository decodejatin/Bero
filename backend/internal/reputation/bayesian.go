// Package reputation implements Bayesian reputation scoring for the Bero platform.
//
// §6.1 — Bayesian Reputation (Beta Distribution + Wilson Score)
//
//	Worker "true quality" p ~ Beta(α₀+S, β₀+F)
//	  α₀=1, β₀=1  → uniform prior (no knowledge about new workers)
//	  S = number of positive reviews (≥4★)
//	  F = number of negative reviews (≤3★)
//
//	Trust Score = lower bound of Wilson Score Confidence Interval:
//	  p̂ = S / (S+F)  (posterior mean approximation)
//	  score = p̂ - z·√(p̂(1-p̂)/n),  z=1.96 for 95% CI
//
// §6.2 — Bayesian Truth Serum (BTS)
//
//	Reward_i = log( P(Rᵢ) / P_prior(Rᵢ) )
//	  P(Rᵢ)       = empirical frequency of rating Rᵢ among peers for similar tasks
//	  P_prior(Rᵢ) = global prior distribution over rating values
package reputation

import "math"

// --- §6.1: Beta Distribution + Wilson Score Trust Score ----------------------

// BetaPrior holds the parameters for the Beta distribution prior.
// The uniform prior Beta(1,1) encodes "we know nothing" about a new worker.
type BetaPrior struct {
	Alpha0 float64 // α₀ — prior pseudo-successes (default 1.0)
	Beta0  float64 // β₀ — prior pseudo-failures  (default 1.0)
}

// DefaultPrior returns the uniform Beta(1,1) prior.
// This gives new workers a neutral TrustScore of 0.5, making them compete
// fairly while clearly distinguishing them from proven high-volume workers.
func DefaultPrior() BetaPrior {
	return BetaPrior{Alpha0: 1.0, Beta0: 1.0}
}

// ComputeTrustScore computes the lower bound of the Wilson Score confidence
// interval on the Beta posterior — the Trust Score exposed to clients.
//
// Parameters:
//   - successes: number of positive reviews (rating ≥ 4)
//   - failures:  number of negative reviews (rating ≤ 3)
//   - prior:     Beta distribution prior (use DefaultPrior())
//   - z:         confidence level z-score (1.96 = 95%, 1.645 = 90%)
//
// Returns a value in [0, 1] where higher = more trustworthy.
// A new worker with 0 reviews returns prior mean ≈ 0.5.
func ComputeTrustScore(successes, failures int, prior BetaPrior, z float64) float64 {
	// Bayesian posterior parameters
	alpha := prior.Alpha0 + float64(successes)
	beta := prior.Beta0 + float64(failures)

	// Posterior mean (Bayesian estimate of true quality)
	pHat := alpha / (alpha + beta)

	// Total effective observations (posterior sample size)
	n := alpha + beta

	// Wilson Score lower confidence bound
	// Penalizes uncertainty: a worker with 1 review is scored conservatively
	variance := pHat * (1.0 - pHat) / n
	score := pHat - z*math.Sqrt(variance)

	// Clamp to [0, 1]
	return math.Max(0.0, math.Min(1.0, score))
}

// RatingToSuccessFailure converts a 1–5 star rating to (successes, failures).
// We use a threshold of 4: ratings ≥ 4 = success, ratings ≤ 3 = failure.
// This models the Bernoulli trial P(good job) as the latent variable.
func RatingToSuccessFailure(rating int) (success, failure int) {
	if rating >= 4 {
		return 1, 0
	}
	return 0, 1
}

// RecomputeTrustScore recomputes TrustScore given running Bayesian parameters.
// This is the main entry point called by the rating service on each new review.
func RecomputeTrustScore(totalSuccesses, totalFailures int) float64 {
	return ComputeTrustScore(totalSuccesses, totalFailures, DefaultPrior(), 1.96)
}

// PosteriorMean returns the expected value of the Beta posterior — the raw
// Bayesian average, before Wilson Score uncertainty penalization.
// Useful for display purposes alongside the conservative TrustScore.
func PosteriorMean(successes, failures int, prior BetaPrior) float64 {
	alpha := prior.Alpha0 + float64(successes)
	beta := prior.Beta0 + float64(failures)
	return alpha / (alpha + beta)
}

// --- §6.2: Bayesian Truth Serum (BTS) ----------------------------------------

// GlobalRatingPrior is the platform-wide prior over 1–5 star ratings.
// Represents the expected distribution of ratings before observing any reviews.
// Derived from industry data: most ratings skew 4-5★ on service platforms.
var GlobalRatingPrior = map[int]float64{
	1: 0.05,
	2: 0.05,
	3: 0.10,
	4: 0.35,
	5: 0.45,
}

// BTSReward computes the Bayesian Truth Serum reward for a single reviewer.
//
// The reward incentivizes honest reporting by comparing the reviewer's signal
// to peer consensus. A reviewer who identifies a "surprising but corroborated"
// flaw (e.g., subtle quality issue others also spotted) is rewarded positively.
// A grade-inflating reviewer (gives 5★ when the consensus is lower) is penalized.
//
// Formula: Reward_i = log( P(Rᵢ | peers) / P_prior(Rᵢ) )
//
// Parameters:
//   - reviewerRating: the rating submitted by this reviewer (1–5)
//   - peerRatings:    all other ratings for the same worker in this category
//   - prior:          global rating prior (use GlobalRatingPrior)
//
// Returns a reward score. Positive = truthful, negative = suspicious.
func BTSReward(reviewerRating int, peerRatings []int, prior map[int]float64) float64 {
	if len(peerRatings) == 0 {
		// Cannot score without peer data; return neutral
		return 0.0
	}

	// Compute empirical peer distribution P(R | peers)
	peerCounts := make(map[int]int)
	for _, r := range peerRatings {
		peerCounts[r]++
	}

	// Laplace-smooth the peer distribution to avoid log(0)
	const smoothing = 1
	numBuckets := 5 // ratings 1–5
	peerProb := make(map[int]float64)
	for r := 1; r <= 5; r++ {
		peerProb[r] = float64(peerCounts[r]+smoothing) / float64(len(peerRatings)+smoothing*numBuckets)
	}

	// P(Rᵢ | peers)
	pObserved := peerProb[reviewerRating]

	// P_prior(Rᵢ) — platform-wide expectation
	pPrior, ok := prior[reviewerRating]
	if !ok || pPrior <= 0 {
		pPrior = 1.0 / float64(numBuckets) // uniform fallback
	}

	// Reward = log( P(Rᵢ | peers) / P_prior(Rᵢ) )
	return math.Log(pObserved / pPrior)
}

// BTSRewardNormalized returns a normalized [−1, 1] version of the BTS reward,
// suitable for immediate display or storage as a "review trust score".
// Uses a logistic function: normalized = 2·σ(reward) − 1
func BTSRewardNormalized(reward float64) float64 {
	sigmoid := 1.0 / (1.0 + math.Exp(-reward))
	return 2*sigmoid - 1
}
