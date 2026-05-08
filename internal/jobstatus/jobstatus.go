// Package jobstatus tracks the last known status of each job (success, failure, running, unknown).
package jobstatus

import (
	"sync"
	"time"
)

// Status represents the last known state of a job.
type Status int

const (
	StatusUnknown Status = iota
	StatusRunning
	StatusSuccess
	StatusFailure
)

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusFailure:
		return "failure"
	default:
		return "unknown"
	}
}

// Entry holds the status and the timestamp of the last transition.
type Entry struct {
	Job       string    `json:"job"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store tracks per-job status.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns a new Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Set records the current status for a job.
func (s *Store) Set(job string, status Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = Entry{
		Job:       job,
		Status:    status.String(),
		UpdatedAt: s.now(),
	}
}

// Get returns the Entry for a job. If the job is unknown the returned bool is false.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// All returns a snapshot of every tracked job.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
