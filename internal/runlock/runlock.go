// Package runlock prevents concurrent execution of the same cron job.
// If a job is already running when its next tick fires, the new execution
// is skipped and the skip is recorded for observability.
package runlock

import (
	"sync"
	"time"
)

// Store tracks which jobs are currently executing.
type Store struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
}

type lockEntry struct {
	running   bool
	startedAt time.Time
	skips     int
}

// New returns an initialised Store.
func New() *Store {
	return &Store{locks: make(map[string]*lockEntry)}
}

// Acquire attempts to mark a job as running. It returns true when the lock
// is acquired and false when the job is already in flight.
func (s *Store) Acquire(job string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.locks[job]
	if ok && e.running {
		e.skips++
		return false
	}
	s.locks[job] = &lockEntry{running: true, startedAt: time.Now()}
	return true
}

// Release marks a job as no longer running.
func (s *Store) Release(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.locks[job]; ok {
		e.running = false
	}
}

// Status returns a snapshot for a single job. The second return value
// is false when the job has never been seen.
func (s *Store) Status(job string) (LockStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.locks[job]
	if !ok {
		return LockStatus{}, false
	}
	return LockStatus{
		Job:       job,
		Running:   e.running,
		StartedAt: e.startedAt,
		Skips:     e.skips,
	}, true
}

// All returns snapshots for every tracked job.
func (s *Store) All() []LockStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]LockStatus, 0, len(s.locks))
	for job, e := range s.locks {
		out = append(out, LockStatus{
			Job:       job,
			Running:   e.running,
			StartedAt: e.startedAt,
			Skips:     e.skips,
		})
	}
	return out
}

// LockStatus is a point-in-time snapshot of a single job's lock state.
type LockStatus struct {
	Job       string    `json:"job"`
	Running   bool      `json:"running"`
	StartedAt time.Time `json:"started_at,omitempty"`
	Skips     int       `json:"skips"`
}
