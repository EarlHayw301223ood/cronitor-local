// Package jobtimeout tracks per-job timeout overrides, allowing individual
// jobs to specify a custom execution deadline that supersedes the global
// runner default.
package jobtimeout

import (
	"errors"
	"sync"
	"time"
)

// ErrInvalidTimeout is returned when a non-positive duration is set.
var ErrInvalidTimeout = errors.New("timeout must be greater than zero")

// Entry holds the configured timeout for a single job.
type Entry struct {
	Job     string        `json:"job"`
	Timeout time.Duration `json:"timeout_ms"`
}

// Store holds per-job timeout overrides.
type Store struct {
	mu       sync.RWMutex
	timeouts map[string]time.Duration
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		timeouts: make(map[string]time.Duration),
	}
}

// Set registers a timeout override for the named job.
// Returns ErrInvalidTimeout if d is not positive.
func (s *Store) Set(job string, d time.Duration) error {
	if d <= 0 {
		return ErrInvalidTimeout
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeouts[job] = d
	return nil
}

// Get returns the configured timeout for job and true, or 0 and false if
// no override has been registered.
func (s *Store) Get(job string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.timeouts[job]
	return d, ok
}

// Delete removes any override for the named job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.timeouts, job)
}

// All returns a snapshot of every registered timeout as a slice of Entry.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.timeouts))
	for job, d := range s.timeouts {
		out = append(out, Entry{Job: job, Timeout: d})
	}
	return out
}
