// Package history records per-job execution history and provides
// a rolling window of recent run results for inspection and alerting.
package history

import (
	"sync"
	"time"
)

// Entry represents a single job execution record.
type Entry struct {
	JobName   string
	StartedAt time.Time
	Duration  time.Duration
	Success   bool
	Output    string
}

// History stores a bounded ring-buffer of execution entries per job.
type History struct {
	mu      sync.RWMutex
	entries map[string][]Entry
	limit   int
}

// New creates a History that retains at most limit entries per job.
func New(limit int) *History {
	if limit <= 0 {
		limit = 50
	}
	return &History{
		entries: make(map[string][]Entry),
		limit:   limit,
	}
}

// Record appends an entry for the given job, evicting the oldest if necessary.
func (h *History) Record(e Entry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	buf := h.entries[e.JobName]
	buf = append(buf, e)
	if len(buf) > h.limit {
		buf = buf[len(buf)-h.limit:]
	}
	h.entries[e.JobName] = buf
}

// Get returns a copy of all recorded entries for the given job.
func (h *History) Get(jobName string) []Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	src := h.entries[jobName]
	out := make([]Entry, len(src))
	copy(out, src)
	return out
}

// All returns a snapshot of entries for every known job.
func (h *History) All() map[string][]Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[string][]Entry, len(h.entries))
	for k, v := range h.entries {
		cp := make([]Entry, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}
