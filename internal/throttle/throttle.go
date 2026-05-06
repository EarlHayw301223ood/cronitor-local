// Package throttle limits how frequently a given job may be dispatched,
// preventing burst re-runs when the scheduler ticks faster than a job's
// expected cadence (e.g. during clock skew recovery).
package throttle

import (
	"sync"
	"time"
)

// Store tracks the last dispatch time for each job.
type Store struct {
	mu      sync.Mutex
	last    map[string]time.Time
	minGap  time.Duration
	nowFunc func() time.Time
}

// New returns a Store that enforces a minimum gap between consecutive
// dispatches of the same job.
func New(minGap time.Duration) *Store {
	return &Store{
		last:    make(map[string]time.Time),
		minGap:  minGap,
		nowFunc: time.Now,
	}
}

// Allow returns true when the job may be dispatched. If the minimum gap
// since the last dispatch has not yet elapsed, it returns false.
func (s *Store) Allow(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFunc()
	if t, ok := s.last[job]; ok && now.Sub(t) < s.minGap {
		return false
	}
	s.last[job] = now
	return true
}

// Reset clears the dispatch record for a job, allowing it to run immediately
// on the next tick regardless of the gap.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.last, job)
}

// LastDispatch returns the time of the most recent allowed dispatch for job,
// and false if the job has never been dispatched.
func (s *Store) LastDispatch(job string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.last[job]
	return t, ok
}

// Snapshot returns a copy of all recorded dispatch times keyed by job name.
func (s *Store) Snapshot() map[string]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]time.Time, len(s.last))
	for k, v := range s.last {
		out[k] = v
	}
	return out
}
