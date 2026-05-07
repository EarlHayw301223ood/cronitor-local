// Package pauseguard provides a mechanism to temporarily pause job execution.
// Jobs can be paused for a fixed duration or indefinitely, and the watcher
// will skip execution of any paused job until the pause expires or is lifted.
package pauseguard

import (
	"sync"
	"time"
)

// Store tracks paused jobs and their expiry times.
type Store struct {
	mu      sync.RWMutex
	pauses  map[string]time.Time // zero Time means indefinite
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		pauses: make(map[string]time.Time),
		now:    time.Now,
	}
}

// Pause suspends job execution for the given duration.
// Pass 0 to pause indefinitely.
func (s *Store) Pause(job string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d == 0 {
		s.pauses[job] = time.Time{}
	} else {
		s.pauses[job] = s.now().Add(d)
	}
}

// Resume lifts the pause on a job immediately.
func (s *Store) Resume(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pauses, job)
}

// IsPaused reports whether the named job is currently paused.
func (s *Store) IsPaused(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, ok := s.pauses[job]
	if !ok {
		return false
	}
	// zero expiry means indefinite pause
	if expiry.IsZero() {
		return true
	}
	return s.now().Before(expiry)
}

// Status holds observable state for a single job's pause.
type Status struct {
	Job       string     `json:"job"`
	Paused    bool       `json:"paused"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// All returns a snapshot of all tracked jobs, including those whose pause
// has already expired.
func (s *Store) All() []Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Status, 0, len(s.pauses))
	now := s.now()
	for job, expiry := range s.pauses {
		st := Status{Job: job}
		if expiry.IsZero() {
			st.Paused = true
		} else if now.Before(expiry) {
			st.Paused = true
			st.ExpiresAt = &expiry
		}
		out = append(out, st)
	}
	return out
}
