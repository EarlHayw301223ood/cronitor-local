// Package execlog provides a bounded, in-memory store for capturing the
// stdout/stderr output of individual job executions.
package execlog

import (
	"sync"
	"time"
)

// Entry holds the captured output of a single job run.
type Entry struct {
	JobName   string    `json:"job_name"`
	StartedAt time.Time `json:"started_at"`
	Output    string    `json:"output"`
	ExitCode  int       `json:"exit_code"`
}

// Store keeps the last N execution log entries per job.
type Store struct {
	mu      sync.RWMutex
	entries map[string][]Entry
	limit   int
}

// New creates a Store that retains at most limit entries per job.
func New(limit int) *Store {
	if limit <= 0 {
		limit = 10
	}
	return &Store{
		entries: make(map[string][]Entry),
		limit:   limit,
	}
}

// Record appends an entry for the given job, evicting the oldest when the
// per-job limit is reached.
func (s *Store) Record(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := s.entries[e.JobName]
	buf = append(buf, e)
	if len(buf) > s.limit {
		buf = buf[len(buf)-s.limit:]
	}
	s.entries[e.JobName] = buf
}

// Get returns all stored entries for the named job, oldest first.
// Returns an empty slice when the job is unknown.
func (s *Store) Get(jobName string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buf := s.entries[jobName]
	out := make([]Entry, len(buf))
	copy(out, buf)
	return out
}

// All returns every entry across all jobs.
func (s *Store) All() map[string][]Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string][]Entry, len(s.entries))
	for k, v := range s.entries {
		cp := make([]Entry, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}
