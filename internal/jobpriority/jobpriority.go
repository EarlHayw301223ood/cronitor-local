// Package jobpriority assigns and tracks execution priority levels for jobs.
// Priority influences scheduling order when multiple jobs are due simultaneously.
package jobpriority

import (
	"fmt"
	"sync"
)

// Level represents a job's execution priority.
type Level int

const (
	Low    Level = 1
	Normal Level = 5
	High   Level = 10
)

// Entry holds the priority configuration for a single job.
type Entry struct {
	Job      string `json:"job"`
	Priority Level  `json:"priority"`
}

// Store records and retrieves priority levels per job.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Level
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Level)}
}

// Set assigns a priority level to the named job.
// Returns an error if the level is not one of Low, Normal or High.
func (s *Store) Set(job string, level Level) error {
	switch level {
	case Low, Normal, High:
	default:
		return fmt.Errorf("jobpriority: unknown level %d", level)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = level
	return nil
}

// Get returns the priority for the named job.
// If no priority has been set, Normal is returned along with false.
func (s *Store) Get(job string) (Level, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.entries[job]
	if !ok {
		return Normal, false
	}
	return l, true
}

// All returns a snapshot of every tracked job and its priority.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for job, lvl := range s.entries {
		out = append(out, Entry{Job: job, Priority: lvl})
	}
	return out
}

// Reset removes the priority record for the named job.
func (s *Store) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}
