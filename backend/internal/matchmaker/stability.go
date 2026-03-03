package matchmaker

import (
	"math"
	"time"
)

// BlockingPair represents a worker-job pair that would both prefer to be matched
// with each other over their current assignment — a threat to matching stability.
type BlockingPair struct {
	WorkerIdx  int     // index into workers slice
	JobIdx     int     // index into jobs slice
	WorkerGain float64 // Uw(j) - Uw(current) - C_switch
	JobGain    float64 // Uj(w) - Uj(current)
}

// TimeDependentUtility computes the time-dependent utility U(t) from Equation (4):
//
//	U(t) = V_job · e^(-λt) − C_travel
//
// Parameters:
//   - vjob: base value of the job to the worker (composite weight from ComputeWeight)
//   - lambda: time decay rate λ (higher = faster decay)
//   - tHours: elapsed time in hours since job creation
//   - cTravel: travel cost (derived from distance / proximity penalty)
func TimeDependentUtility(vjob, lambda, tHours, cTravel float64) float64 {
	return vjob*math.Exp(-lambda*tHours) - cTravel
}

// workerUtilityForJob computes the time-dependent utility a worker derives from a job.
// It combines the composite weight (as V_job) with exponential time decay and travel cost.
func workerUtilityForJob(worker MatchableWorker, job MatchableJob, cfg MatchConfig, now time.Time) float64 {
	vjob := ComputeWeight(worker, job, cfg, now)
	tHours := now.Sub(job.CreatedAt).Hours()
	if tHours < 0 {
		tHours = 0
	}

	// Travel cost: inverse of proximity score (farther = higher cost)
	proximity := computeProximity(worker, job, cfg)
	cTravel := (1.0 - proximity) * cfg.AlphaProximity

	return TimeDependentUtility(vjob, cfg.LambdaDecay, tHours, cTravel)
}

// jobUtilityForWorker computes the utility a job derives from being assigned to a worker.
// From the job's perspective, utility is based on skill match, reputation, and proximity.
func jobUtilityForWorker(worker MatchableWorker, job MatchableJob, cfg MatchConfig, now time.Time) float64 {
	// The job's preference is the same composite weight — the platform acts as the
	// job-side agent and optimizes for the same objective function.
	return ComputeWeight(worker, job, cfg, now)
}

// FindBlockingPairs identifies all blocking pairs in a matching according to the
// modified stability condition from Equation (5):
//
//	(Uw(j,t) > Uw(µ(w),t) + C_switch) ∧ (Uj(w,t) > Uj(µ(j),t))
//
// A matching is stable if this function returns an empty slice.
func FindBlockingPairs(
	assignments []Assignment,
	workers []MatchableWorker,
	jobs []MatchableJob,
	cfg MatchConfig,
	now time.Time,
) []BlockingPair {
	// Build lookup maps: workerID → assigned jobIdx, jobID → assigned workerIdx
	workerToJob := make(map[string]int) // workerID → jobIdx (or -1)
	jobToWorker := make(map[string]int) // jobID → workerIdx (or -1)

	// Build ID → index maps
	workerIDToIdx := make(map[string]int, len(workers))
	for i, w := range workers {
		workerIDToIdx[w.ID] = i
	}
	jobIDToIdx := make(map[string]int, len(jobs))
	for j, job := range jobs {
		jobIDToIdx[job.ID] = j
	}

	// Initialize all as unassigned (-1)
	for _, w := range workers {
		workerToJob[w.ID] = -1
	}
	for _, j := range jobs {
		jobToWorker[j.ID] = -1
	}

	// Record current assignments
	for _, a := range assignments {
		if jIdx, ok := jobIDToIdx[a.JobID]; ok {
			workerToJob[a.WorkerID] = jIdx
		}
		if wIdx, ok := workerIDToIdx[a.WorkerID]; ok {
			jobToWorker[a.JobID] = wIdx
		}
	}

	var blockingPairs []BlockingPair

	for wi, w := range workers {
		currentJobIdx := workerToJob[w.ID]

		// Worker's current utility (0 if unassigned)
		var uwCurrent float64
		if currentJobIdx >= 0 {
			uwCurrent = workerUtilityForJob(w, jobs[currentJobIdx], cfg, now)
		}

		for ji, j := range jobs {
			// Skip if already assigned to this job
			if ji == currentJobIdx {
				continue
			}

			// Worker's utility for the candidate job
			uwCandidate := workerUtilityForJob(w, j, cfg, now)

			// Check worker's blocking condition: Uw(j) > Uw(µ(w)) + C_switch
			workerGain := uwCandidate - uwCurrent - cfg.SwitchingCost
			if workerGain <= 0 {
				continue
			}

			// Job's current utility (0 if unassigned)
			currentWorkerIdx := jobToWorker[j.ID]
			var ujCurrent float64
			if currentWorkerIdx >= 0 {
				ujCurrent = jobUtilityForWorker(workers[currentWorkerIdx], j, cfg, now)
			}

			// Job's utility for the candidate worker
			ujCandidate := jobUtilityForWorker(w, j, cfg, now)

			// Check job's blocking condition: Uj(w) > Uj(µ(j))
			jobGain := ujCandidate - ujCurrent
			if jobGain <= 0 {
				continue
			}

			// Both conditions met — this is a blocking pair
			blockingPairs = append(blockingPairs, BlockingPair{
				WorkerIdx:  wi,
				JobIdx:     ji,
				WorkerGain: workerGain,
				JobGain:    jobGain,
			})
		}
	}

	return blockingPairs
}

// ResolveBlockingPairs uses a modified Deferred Acceptance (Gale-Shapley) approach
// to resolve blocking pairs and produce a stable matching.
//
// Workers act as proposers: each blocking worker proposes to their most-preferred
// job. Jobs tentatively accept the best proposer if they prefer them over the
// current holder (accounting for switching costs). Rejected workers propose to
// their next best option. The algorithm terminates when no blocking pairs remain.
func ResolveBlockingPairs(
	assignments []Assignment,
	workers []MatchableWorker,
	jobs []MatchableJob,
	cfg MatchConfig,
	now time.Time,
) []Assignment {
	nWorkers := len(workers)
	nJobs := len(jobs)

	// Build ID → index maps
	workerIDToIdx := make(map[string]int, nWorkers)
	for i, w := range workers {
		workerIDToIdx[w.ID] = i
	}
	jobIDToIdx := make(map[string]int, nJobs)
	for j, job := range jobs {
		jobIDToIdx[job.ID] = j
	}

	// Current assignment state: workerAssign[wi] = ji, jobAssign[ji] = wi (-1 = free)
	workerAssign := make([]int, nWorkers)
	jobAssign := make([]int, nJobs)
	for i := range workerAssign {
		workerAssign[i] = -1
	}
	for j := range jobAssign {
		jobAssign[j] = -1
	}

	// Initialize from existing assignments
	for _, a := range assignments {
		wi, wOk := workerIDToIdx[a.WorkerID]
		ji, jOk := jobIDToIdx[a.JobID]
		if wOk && jOk {
			workerAssign[wi] = ji
			jobAssign[ji] = wi
		}
	}

	// Pre-compute worker utility for each job
	wUtil := make([][]float64, nWorkers)
	for wi := range workers {
		wUtil[wi] = make([]float64, nJobs)
		for ji := range jobs {
			wUtil[wi][ji] = workerUtilityForJob(workers[wi], jobs[ji], cfg, now)
		}
	}

	// Pre-compute job utility for each worker
	jUtil := make([][]float64, nJobs)
	for ji := range jobs {
		jUtil[ji] = make([]float64, nWorkers)
		for wi := range workers {
			jUtil[ji][wi] = jobUtilityForWorker(workers[wi], jobs[ji], cfg, now)
		}
	}

	// Build worker preference lists: jobs sorted by decreasing utility
	workerPrefList := make([][]int, nWorkers)
	for wi := range workers {
		prefs := make([]int, nJobs)
		for ji := range jobs {
			prefs[ji] = ji
		}
		wiCopy := wi
		sortByDesc(prefs, func(ji int) float64 { return wUtil[wiCopy][ji] })
		workerPrefList[wi] = prefs
	}

	// Track proposal progress: nextProposal[wi] = index into workerPrefList[wi]
	nextProposal := make([]int, nWorkers)

	// Identify initially free/blocking workers who want to switch
	freeWorkers := make([]int, 0)
	for wi := range workers {
		if workerAssign[wi] == -1 {
			freeWorkers = append(freeWorkers, wi)
		} else {
			// Check if this worker is in a blocking pair (wants to switch)
			currentUtil := wUtil[wi][workerAssign[wi]]
			for _, ji := range workerPrefList[wi] {
				if ji == workerAssign[wi] {
					continue
				}
				if wUtil[wi][ji] > currentUtil+cfg.SwitchingCost {
					// Worker wants to switch — add to proposers
					freeWorkers = append(freeWorkers, wi)
					break
				}
			}
		}
	}

	// Deferred Acceptance loop
	maxIterations := nWorkers * nJobs // safety bound
	for iter := 0; iter < maxIterations && len(freeWorkers) > 0; iter++ {
		wi := freeWorkers[0]
		freeWorkers = freeWorkers[1:]

		// Skip if already exhausted all proposals
		if nextProposal[wi] >= nJobs {
			continue
		}

		// Find the next job to propose to
		ji := workerPrefList[wi][nextProposal[wi]]
		nextProposal[wi]++

		candidateUtil := wUtil[wi][ji]

		// Worker must gain enough to overcome switching cost (if already assigned)
		if workerAssign[wi] >= 0 {
			currentUtil := wUtil[wi][workerAssign[wi]]
			if candidateUtil <= currentUtil+cfg.SwitchingCost {
				// Not worth switching, try next
				if nextProposal[wi] < nJobs {
					freeWorkers = append(freeWorkers, wi)
				}
				continue
			}
		}

		// Check if job is free
		if jobAssign[ji] == -1 {
			// Job accepts
			oldJob := workerAssign[wi]
			if oldJob >= 0 {
				jobAssign[oldJob] = -1 // free old job
			}
			workerAssign[wi] = ji
			jobAssign[ji] = wi
		} else {
			// Job compares current holder vs proposer
			currentHolder := jobAssign[ji]
			if jUtil[ji][wi] > jUtil[ji][currentHolder] {
				// Job prefers proposer — accept and reject current holder
				oldJob := workerAssign[wi]
				if oldJob >= 0 {
					jobAssign[oldJob] = -1
				}
				workerAssign[wi] = ji
				jobAssign[ji] = wi

				// Current holder becomes free and can re-propose
				workerAssign[currentHolder] = -1
				freeWorkers = append(freeWorkers, currentHolder)
			} else {
				// Job rejects — worker tries next preference
				if nextProposal[wi] < nJobs {
					freeWorkers = append(freeWorkers, wi)
				}
			}
		}
	}

	// Build final assignment list
	var result []Assignment
	for wi, ji := range workerAssign {
		if ji >= 0 {
			w := ComputeWeight(workers[wi], jobs[ji], cfg, now)
			if w >= cfg.MinWeightThreshold {
				result = append(result, Assignment{
					WorkerID: workers[wi].ID,
					JobID:    jobs[ji].ID,
					Weight:   w,
				})
			}
		}
	}

	return result
}

// EnforceStability is the public entry point for stability enforcement.
// Given a set of assignments from the Hungarian algorithm, it checks for
// blocking pairs and resolves them using modified Deferred Acceptance.
//
// If no blocking pairs exist, the original assignments are returned unchanged.
func EnforceStability(
	assignments []Assignment,
	workers []MatchableWorker,
	jobs []MatchableJob,
	cfg MatchConfig,
	now time.Time,
) []Assignment {
	// Fast path: check if matching is already stable
	blockingPairs := FindBlockingPairs(assignments, workers, jobs, cfg, now)
	if len(blockingPairs) == 0 {
		return assignments
	}

	// Resolve via Deferred Acceptance
	return ResolveBlockingPairs(assignments, workers, jobs, cfg, now)
}

// sortByDesc sorts indices in descending order of the value returned by fn.
func sortByDesc(indices []int, fn func(int) float64) {
	// Simple insertion sort (preference lists are typically small, k ≤ 10-20)
	for i := 1; i < len(indices); i++ {
		key := indices[i]
		keyVal := fn(key)
		j := i - 1
		for j >= 0 && fn(indices[j]) < keyVal {
			indices[j+1] = indices[j]
			j--
		}
		indices[j+1] = key
	}
}
