// Package jobenv stores per-job environment variable overrides that are
// injected into the process environment when a job is executed.
package jobenv

import (
	"fmt"
	"sync"
)

// Store holds environment variable maps keyed by job name.
type Store struct {
	mu   sync.RWMutex
	envs map[string]map[string]string
}

// New returns an initialised Store.
func New() *Store {
	return &Store{envs: make(map[string]map[string]string)}
}

// Set replaces the entire environment map for the named job.
// An empty map is valid and clears all overrides for the job.
func (s *Store) Set(job string, env map[string]string) error {
	if job == "" {
		return fmt.Errorf("jobenv: job name must not be empty")
	}
	copy := make(map[string]string, len(env))
	for k, v := range env {
		copy[k] = v
	}
	s.mu.Lock()
	s.envs[job] = copy
	s.mu.Unlock()
	return nil
}

// Get returns the environment map for the named job and whether it was found.
func (s *Store) Get(job string) (map[string]string, bool) {
	s.mu.RLock()
	env, ok := s.envs[job]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	copy := make(map[string]string, len(env))
	for k, v := range env {
		copy[k] = v
	}
	return copy, true
}

// Delete removes all environment overrides for the named job.
func (s *Store) Delete(job string) {
	s.mu.Lock()
	delete(s.envs, job)
	s.mu.Unlock()
}

// All returns a snapshot of every job's environment map.
func (s *Store) All() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]string, len(s.envs))
	for job, env := range s.envs {
		copy := make(map[string]string, len(env))
		for k, v := range env {
			copy[k] = v
		}
		out[job] = copy
	}
	return out
}
