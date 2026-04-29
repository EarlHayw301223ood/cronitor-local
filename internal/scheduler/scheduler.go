package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a scheduled command.
type Job struct {
	Name     string
	Command  string
	Schedule string
}

type status struct {
	LastRun    time.Time
	LastError  error
	RunCount   int
	ErrorCount int
}

// Scheduler tracks registered jobs and their execution status.
type Scheduler struct {
	mu       sync.RWMutex
	jobs     map[string]Job
	statuses map[string]*status
	parser   cron.Parser
}

// New creates an empty Scheduler.
func New() *Scheduler {
	return &Scheduler{
		jobs:     make(map[string]Job),
		statuses: make(map[string]*status),
		parser:   cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// Register adds a job after validating its cron schedule.
func (s *Scheduler) Register(j Job) error {
	if _, err := s.parser.Parse(j.Schedule); err != nil {
		return fmt.Errorf("invalid schedule %q for job %q: %w", j.Schedule, j.Name, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.Name] = j
	s.statuses[j.Name] = &status{}
	return nil
}

// Due returns jobs whose schedule matches the given time (minute precision).
func (s *Scheduler) Due(now time.Time) []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now = now.Truncate(time.Minute)
	var due []Job
	for _, j := range s.jobs {
		sched, err := s.parser.Parse(j.Schedule)
		if err != nil {
			continue
		}
		if sched.Next(now.Add(-time.Minute)).Equal(now) {
			due = append(due, j)
		}
	}
	return due
}

// RecordSuccess marks a successful run for the named job.
func (s *Scheduler) RecordSuccess(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.statuses[name]; ok {
		st.LastRun = time.Now()
		st.RunCount++
		st.LastError = nil
	}
}

// RecordError marks a failed run for the named job.
func (s *Scheduler) RecordError(name string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.statuses[name]; ok {
		st.LastRun = time.Now()
		st.RunCount++
		st.ErrorCount++
		st.LastError = err
	}
}

// Status returns a snapshot of a job's run history.
func (s *Scheduler) Status(name string) (runCount, errCount int, lastErr error, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.statuses[name]
	if !ok {
		return
	}
	return st.RunCount, st.ErrorCount, st.LastError, true
}
