// Package execcount tracks the total number of executions per job,
// including success and failure counts, for diagnostic and reporting purposes.
package execcount

import "sync"

// Store holds execution counters for each job.
type Store struct {
	mu      sync.RWMutex
	counts  map[string]*Counter
}

// Counter holds the execution statistics for a single job.
type Counter struct {
	Total    int64 `json:"total"`
	Success  int64 `json:"success"`
	Failures int64 `json:"failures"`
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		counts: make(map[string]*Counter),
	}
}

// Record increments the counters for the named job.
// ok should be true for a successful execution and false for a failure.
func (s *Store) Record(job string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.counts[job]
	if !exists {
		c = &Counter{}
		s.counts[job] = c
	}

	c.Total++
	if ok {
		c.Success++
	} else {
		c.Failures++
	}
}

// Get returns the Counter for the named job.
// A zero-value Counter is returned if the job has never been recorded.
func (s *Store) Get(job string) Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.counts[job]; ok {
		return *c
	}
	return Counter{}
}

// All returns a snapshot of counters for every tracked job.
func (s *Store) All() map[string]Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]Counter, len(s.counts))
	for k, v := range s.counts {
		out[k] = *v
	}
	return out
}

// Reset clears all counters for the named job.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.counts, job)
}
