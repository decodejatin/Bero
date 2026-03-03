package reputation

import (
	"math"
	"testing"
)

// --- §6.1: TrustScore tests ---

func TestComputeTrustScore_ColdStart(t *testing.T) {
	// New worker: 0 successes, 0 failures
	// With Beta(1,1) prior: p̂ = 0.5, n = 2 (prior pseudo-counts)
	// Wilson Score = 0.5 - 1.96·√(0.25/2) = -0.19 → clamped to 0.0
	// This is correct: new workers score 0 until they earn real reviews.
	// The fallback in computeReputation returns 0.5 for them as a neutral start.
	score := RecomputeTrustScore(0, 0)
	t.Logf("Cold start TrustScore: %.4f", score)
	if score < 0 || score > 0.2 {
		t.Errorf("expected cold start score in [0, 0.2] (conservative Wilson bound), got %.4f", score)
	}
}

func TestComputeTrustScore_UncertaintyPenalty(t *testing.T) {
	// One 5★ vs a hundred 5★ — the latter should score much higher
	// because Wilson CI penalizes small sample size
	one5Star := RecomputeTrustScore(1, 0)
	hundred5Star := RecomputeTrustScore(100, 0)

	t.Logf("1×5★ TrustScore: %.4f", one5Star)
	t.Logf("100×5★ TrustScore: %.4f", hundred5Star)

	if one5Star >= hundred5Star {
		t.Errorf("expected 1×5★ (%.4f) < 100×5★ (%.4f), uncertainty should penalize small sample",
			one5Star, hundred5Star)
	}
	// Sanity: 100×5★ should be very high
	if hundred5Star < 0.90 {
		t.Errorf("expected 100×5★ TrustScore > 0.90, got %.4f", hundred5Star)
	}
}

func TestComputeTrustScore_FailuresPenalized(t *testing.T) {
	// 50×5★ alone vs 50×5★ + 50×1★ — mixed reviews reduce score
	allGood := RecomputeTrustScore(50, 0)
	mixed := RecomputeTrustScore(50, 50)

	t.Logf("50×5★ only: %.4f", allGood)
	t.Logf("50×5★ + 50×1★: %.4f", mixed)

	if mixed >= allGood {
		t.Errorf("expected mixed reviews (%.4f) < all good (%.4f)", mixed, allGood)
	}
}

func TestComputeTrustScore_ZeroRatingAvgWorker(t *testing.T) {
	// Worker with no reviews: Wilson Score is clamped to 0
	// (inherently conservative — requires actual reviews to earn trust)
	score := RecomputeTrustScore(0, 0)
	if score < 0 || score > 1 {
		t.Errorf("score must be in [0,1], got %.4f", score)
	}
	t.Logf("Zero-review worker TrustScore: %.4f (conservative Wilson bound)", score)
}

func TestPosteriorMean(t *testing.T) {
	prior := DefaultPrior()

	// 10 successes, 0 failures → mean close to 1
	mean := PosteriorMean(10, 0, prior)
	t.Logf("10 successes, 0 failures: posterior mean = %.4f", mean)
	if mean < 0.80 {
		t.Errorf("expected posterior mean > 0.80, got %.4f", mean)
	}

	// 0 successes, 10 failures → mean close to 0
	mean = PosteriorMean(0, 10, prior)
	t.Logf("0 successes, 10 failures: posterior mean = %.4f", mean)
	if mean > 0.20 {
		t.Errorf("expected posterior mean < 0.20, got %.4f", mean)
	}
}

func TestRatingToSuccessFailure(t *testing.T) {
	cases := []struct {
		rating      int
		wantSuccess int
		wantFailure int
	}{
		{5, 1, 0},
		{4, 1, 0},
		{3, 0, 1},
		{2, 0, 1},
		{1, 0, 1},
	}
	for _, tc := range cases {
		s, f := RatingToSuccessFailure(tc.rating)
		if s != tc.wantSuccess || f != tc.wantFailure {
			t.Errorf("rating %d → want (S=%d, F=%d), got (S=%d, F=%d)",
				tc.rating, tc.wantSuccess, tc.wantFailure, s, f)
		}
	}
}

func TestComputeTrustScore_ConfidenceIntervalDecreases(t *testing.T) {
	// Increasing z (higher confidence requirement) → lower score
	low := ComputeTrustScore(10, 2, DefaultPrior(), 1.0)
	high := ComputeTrustScore(10, 2, DefaultPrior(), 2.576) // 99% CI

	t.Logf("z=1.0:   %.4f", low)
	t.Logf("z=2.576: %.4f", high)

	if low <= high {
		t.Errorf("higher z should yield lower bound, got low=%.4f high=%.4f", low, high)
	}
}

// --- §6.2: BTS tests ---

func TestBTSReward_HonestReview(t *testing.T) {
	// Reviewer gives 3★ for a worker who mostly gets 3–4★ from peers.
	// This is "surprising but corroborated" → positive reward.
	peerRatings := []int{4, 3, 3, 4, 3, 5, 3, 4}
	reward := BTSReward(3, peerRatings, GlobalRatingPrior)
	t.Logf("Honest 3★ reviewer reward: %.4f", reward)
	// The prior expects very few 3★; this peer set is heavier on 3★ → positive
	if reward < 0 {
		t.Errorf("expected positive BTS reward for corroborated surprising review, got %.4f", reward)
	}
}

func TestBTSReward_GradeInflation(t *testing.T) {
	// Reviewer gives 5★ when peers mostly give 3★.
	// This is aligned with the (inflated) prior for 5★ but NOT with peers → negative.
	peerRatings := []int{3, 2, 3, 3, 2, 3, 2, 3}
	reward := BTSReward(5, peerRatings, GlobalRatingPrior)
	t.Logf("Grade-inflating 5★ reviewer reward: %.4f", reward)
	if reward > 0 {
		t.Errorf("expected negative/zero BTS reward for grade inflation, got %.4f", reward)
	}
}

func TestBTSReward_NoPeers(t *testing.T) {
	// Without peer data, reward is neutral (0)
	reward := BTSReward(5, []int{}, GlobalRatingPrior)
	if reward != 0.0 {
		t.Errorf("expected 0 reward with no peers, got %.4f", reward)
	}
}

func TestBTSReward_PeerAligned(t *testing.T) {
	// Reviewer gives 5★, peers also mostly give 5★.
	// The 5★ is expected by both prior AND peers → near-zero or slightly positive.
	peerRatings := []int{5, 5, 5, 4, 5, 5}
	reward := BTSReward(5, peerRatings, GlobalRatingPrior)
	t.Logf("Peer-aligned 5★ reviewer reward: %.4f", reward)
	// Should be small (not strongly rewarded or penalized)
	if math.Abs(reward) > 2.0 {
		t.Errorf("expected near-zero reward for aligned review, got %.4f", reward)
	}
}

func TestBTSRewardNormalized(t *testing.T) {
	cases := []struct {
		rawReward float64
		wantRange [2]float64
	}{
		{0.0, [2]float64{-0.01, 0.01}},   // 0 → normalized ≈ 0
		{10.0, [2]float64{0.99, 1.0}},    // very positive → near +1
		{-10.0, [2]float64{-1.0, -0.99}}, // very negative → near -1
	}
	for _, tc := range cases {
		n := BTSRewardNormalized(tc.rawReward)
		t.Logf("raw=%.1f → normalized=%.4f", tc.rawReward, n)
		if n < tc.wantRange[0] || n > tc.wantRange[1] {
			t.Errorf("raw %.1f: want [%.2f, %.2f], got %.4f",
				tc.rawReward, tc.wantRange[0], tc.wantRange[1], n)
		}
	}
}
