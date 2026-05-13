// Package jobdependency tracks inter-job dependencies, allowing a job to
// declare that it must not run until one or more upstream jobs have completed
// successfully within a given time window.
package jobdependency

import (
	"fmt"
	"sync"
	"time"
)

// Store holds dependency declarations and the last successful completion
// timestamps used to evaluate readiness.
type Store struct {
	mu           sync.RWMutex
	deps         map[string][]string    // job -> upstream jobs it depends on
	lastSuccess  map[string]time.Time   // job -> last successful completion
	window       time.Duration          // how recent a success must be
	now          func() time.Time
}

// New returns an initialised Store. window is the maximum age of an upstream
// success that still satisfies the dependency.
func New(window time.Duration) *Store {
	return &Store{
		deps:        make(map[string][]string),
		lastSuccess: make(map[string]time.Time),
		window:      window,
		now:         time.Now,
	}
}

// Declare registers that job depends on each of the upstream jobs listed.
// Calling Declare again for the same job replaces its previous dependencies.
func (s *Store) Declare(job string, upstreams []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deps[job] = append([]string(nil), upstreams...)
}

// RecordSuccess marks job as having completed successfully at the current time.
func (s *Store) RecordSuccess(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSuccess[job] = s.now()
}

// Ready reports whether all upstream dependencies for job have a recorded
// success within the configured window. It returns the first blocking upstream
// name and an error if the job is not ready.
func (s *Store) Ready(job string) (bool, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	upstreams, ok := s.deps[job]
	if !ok || len(upstreams) == 0 {
		return true, "", nil
	}

	cutoff := s.now().Add(-s.window)
	for _, up := range upstreams {
		t, seen := s.lastSuccess[up]
		if !seen || t.Before(cutoff) {
			return false, up, fmt.Errorf("upstream %q has not succeeded within %s", up, s.window)
		}
	}
	return true, "", nil
}

// All returns a snapshot of every job's declared upstreams.
func (s *Store) All() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.deps))
	for k, v := range s.deps {
		out[k] = append([]string(nil), v...)
	}
	return out
}
