// Package jobqueue provides a bounded, thread-safe queue for pending job
// executions. Jobs that cannot be dispatched immediately (e.g. because the
// runner is busy) are held here until capacity is available or they are
// evicted due to the queue being full.
package jobqueue

import (
	"sync"
	"time"
)

// Entry holds a single queued job execution request.
type Entry struct {
	JobName   string
	EnqueuedAt time.Time
}

// Queue is a bounded FIFO queue of pending job entries.
type Queue struct {
	mu      sync.Mutex
	items   []Entry
	limit   int
	evicted map[string]int
}

// New creates a Queue with the given maximum capacity.
func New(limit int) *Queue {
	if limit <= 0 {
		limit = 64
	}
	return &Queue{
		limit:   limit,
		evicted: make(map[string]int),
	}
}

// Enqueue adds a job to the tail of the queue. If the queue is full the
// oldest entry is evicted and the eviction counter for that job is
// incremented.
func (q *Queue) Enqueue(jobName string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.limit {
		evicted := q.items[0]
		q.items = q.items[1:]
		q.evicted[evicted.JobName]++
	}

	q.items = append(q.items, Entry{
		JobName:    jobName,
		EnqueuedAt: time.Now(),
	})
	return true
}

// Dequeue removes and returns the oldest entry. The second return value is
// false when the queue is empty.
func (q *Queue) Dequeue() (Entry, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return Entry{}, false
	}
	e := q.items[0]
	q.items = q.items[1:]
	return e, true
}

// Len returns the current number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Evictions returns the number of times entries for a given job were evicted
// due to the queue being full.
func (q *Queue) Evictions(jobName string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.evicted[jobName]
}

// AllEvictions returns a copy of the eviction counters for all jobs.
func (q *Queue) AllEvictions() map[string]int {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make(map[string]int, len(q.evicted))
	for k, v := range q.evicted {
		out[k] = v
	}
	return out
}
