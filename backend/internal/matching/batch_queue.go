package matching

import (
	"log"
	"sync"
	"time"
)

// =============================================================================
// BatchQueue — Sliding window batch collector for the matching engine.
//
// Architecture:
//   - Goroutine ticker fires every BatchInterval (default 20s)
//   - Jobs are enqueued via non-blocking Enqueue()
//   - On tick: drains up to MaxBatchSize jobs → invokes ProcessFunc
//   - Excess jobs roll to the next window
//   - Thread-safe via sync.Mutex
// =============================================================================

const (
	DefaultBatchInterval = 20 * time.Second
	DefaultMaxBatchSize  = 30
)

// ProcessFunc is called when a batch is ready.
type ProcessFunc func(jobIDs []string)

// BatchQueue collects jobs and processes them in batches.
type BatchQueue struct {
	mu            sync.Mutex
	queue         []string
	processFunc   ProcessFunc
	batchInterval time.Duration
	maxBatchSize  int
	ticker        *time.Ticker
	stopCh        chan struct{}
	running       bool

	// Stats
	lastBatchAt  *time.Time
	isProcessing bool
}

// NewBatchQueue creates a new batch queue.
func NewBatchQueue(processFunc ProcessFunc) *BatchQueue {
	return &BatchQueue{
		queue:         make([]string, 0, DefaultMaxBatchSize),
		processFunc:   processFunc,
		batchInterval: DefaultBatchInterval,
		maxBatchSize:  DefaultMaxBatchSize,
		stopCh:        make(chan struct{}),
	}
}

// Enqueue adds a job ID to the batch queue. Non-blocking.
func (q *BatchQueue) Enqueue(jobID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Deduplicate
	for _, id := range q.queue {
		if id == jobID {
			return
		}
	}

	q.queue = append(q.queue, jobID)
	log.Printf("[BatchQueue] Enqueued job %s (queue depth: %d)", jobID, len(q.queue))
}

// EnqueueMultiple adds multiple job IDs (used for requeuing failed jobs).
func (q *BatchQueue) EnqueueMultiple(jobIDs []string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	existing := make(map[string]bool, len(q.queue))
	for _, id := range q.queue {
		existing[id] = true
	}

	for _, id := range jobIDs {
		if !existing[id] {
			q.queue = append(q.queue, id)
			existing[id] = true
		}
	}
	log.Printf("[BatchQueue] Enqueued %d jobs (queue depth: %d)", len(jobIDs), len(q.queue))
}

// Drain removes up to maxBatchSize jobs from the queue and returns them.
func (q *BatchQueue) Drain() []string {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	size := len(q.queue)
	if size > q.maxBatchSize {
		size = q.maxBatchSize
	}

	batch := make([]string, size)
	copy(batch, q.queue[:size])
	q.queue = q.queue[size:] // Remainder stays for next window

	return batch
}

// Start begins the batch processing goroutine.
func (q *BatchQueue) Start() {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return
	}
	q.running = true
	q.mu.Unlock()

	q.ticker = time.NewTicker(q.batchInterval)
	log.Printf("[BatchQueue] Started (interval: %v, max batch: %d)", q.batchInterval, q.maxBatchSize)

	go func() {
		for {
			select {
			case <-q.ticker.C:
				q.processTick()
			case <-q.stopCh:
				q.ticker.Stop()
				log.Println("[BatchQueue] Stopped")
				return
			}
		}
	}()
}

// Stop halts the batch processing goroutine.
func (q *BatchQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}
	q.running = false
	close(q.stopCh)
}

// TriggerNow forces an immediate batch processing cycle.
func (q *BatchQueue) TriggerNow() {
	go q.processTick()
}

// QueueDepth returns the current number of jobs waiting.
func (q *BatchQueue) QueueDepth() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

// LastBatchAt returns when the last batch was processed.
func (q *BatchQueue) LastBatchAt() *time.Time {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.lastBatchAt
}

// IsProcessing returns whether a batch is currently being processed.
func (q *BatchQueue) IsProcessing() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.isProcessing
}

// IsRunning returns whether the queue scheduler is running.
func (q *BatchQueue) IsRunning() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.running
}

func (q *BatchQueue) processTick() {
	batch := q.Drain()
	if len(batch) == 0 {
		return
	}

	q.mu.Lock()
	q.isProcessing = true
	q.mu.Unlock()

	now := time.Now()
	log.Printf("[BatchQueue] Processing batch of %d jobs", len(batch))

	// Invoke the processor
	q.processFunc(batch)

	q.mu.Lock()
	q.lastBatchAt = &now
	q.isProcessing = false
	q.mu.Unlock()

	log.Printf("[BatchQueue] Batch complete (took %v)", time.Since(now))
}
