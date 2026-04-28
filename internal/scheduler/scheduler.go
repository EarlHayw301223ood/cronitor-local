package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"cronitor-local/internal/config"
)

// JobStatus tracks the last execution state of a cron job.
type JobStatus struct {
	Name      string
	LastRun   time.Time
	LastError error
	RunCount  int
}

// Scheduler wraps the cron scheduler and tracks job statuses.
type Scheduler struct {
	cron    *cron.Cron
	statuses map[string]*JobStatus
	mu      sync.RWMutex
}

// New creates a new Scheduler instance.
func New() *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		statuses: make(map[string]*JobStatus),
	}
}

// Register adds all jobs from the config to the cron scheduler.
func (s *Scheduler) Register(jobs []config.Job, runner func(config.Job) error) error {
	for _, job := range jobs {
		j := job // capture loop variable
		status := &JobStatus{Name: j.Name}
		s.mu.Lock()
		s.statuses[j.Name] = status
		s.mu.Unlock()

		_, err := s.cron.AddFunc(j.Schedule, func() {
			log.Printf("[scheduler] running job: %s", j.Name)
			err := runner(j)

			s.mu.Lock()
			defer s.mu.Unlock()
			status.LastRun = time.Now()
			status.RunCount++
			status.LastError = err

			if err != nil {
				log.Printf("[scheduler] job %s failed: %v", j.Name, err)
			} else {
				log.Printf("[scheduler] job %s completed successfully", j.Name)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Start begins the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("[scheduler] started")
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("[scheduler] stopped")
}

// Status returns a copy of the current job statuses.
func (s *Scheduler) Status() map[string]JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]JobStatus, len(s.statuses))
	for k, v := range s.statuses {
		out[k] = *v
	}
	return out
}
