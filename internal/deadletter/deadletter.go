// Package deadletter stores jobs that have exhausted their retry budget
// so that operators can inspect and optionally requeue them.
package deadletter

import (
	"sync"
	"time"
)

// Entry records a job execution that could not be completed after all retries.
type Entry struct {
	JobName   string    `json:"job_name"`
	Command   string    `json:"command"`
	Error     string    `json:"error"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

// Store holds dead-letter entries up to a configurable limit.
type Store struct {
	mu      sync.RWMutex
	entries []Entry
	limit   int
}

// New creates a Store that retains at most limit entries (oldest dropped first).
func New(limit int) *Store {
	if limit <= 0 {
		limit = 100
	}
	return &Store{limit: limit}
}

// Add appends an entry, evicting the oldest when the limit is reached.
func (s *Store) Add(e Entry) {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= s.limit {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
}

// All returns a snapshot of all stored entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// ForJob returns entries that match the given job name.
func (s *Store) ForJob(name string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Entry
	for _, e := range s.entries {
		if e.JobName == name {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the current number of stored entries.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
