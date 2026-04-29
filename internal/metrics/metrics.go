// Package metrics tracks runtime statistics for monitored cron jobs.
package metrics

import (
	"sync"
	"time"
)

// JobMetrics holds execution statistics for a single job.
type JobMetrics struct {
	Name        string
	TotalRuns   int64
	Failures    int64
	LastRun     time.Time
	LastSuccess time.Time
	LastFailure time.Time
	AvgDuration time.Duration
	totalDur    time.Duration
}

// Collector aggregates metrics across all registered jobs.
type Collector struct {
	mu   sync.RWMutex
	jobs map[string]*JobMetrics
}

// New returns an initialised Collector.
func New() *Collector {
	return &Collector{jobs: make(map[string]*JobMetrics)}
}

// RecordRun records a completed job execution.
func (c *Collector) RecordRun(name string, dur time.Duration, failed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.jobs[name]
	if !ok {
		m = &JobMetrics{Name: name}
		c.jobs[name] = m
	}

	now := time.Now()
	m.TotalRuns++
	m.LastRun = now
	m.totalDur += dur
	m.AvgDuration = m.totalDur / time.Duration(m.TotalRuns)

	if failed {
		m.Failures++
		m.LastFailure = now
	} else {
		m.LastSuccess = now
	}
}

// Snapshot returns a copy of metrics for the named job.
// Returns false if the job has not been seen yet.
func (c *Collector) Snapshot(name string) (JobMetrics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.jobs[name]
	if !ok {
		return JobMetrics{}, false
	}
	return *m, true
}

// All returns a slice of snapshots for every tracked job.
func (c *Collector) All() []JobMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]JobMetrics, 0, len(c.jobs))
	for _, m := range c.jobs {
		out = append(out, *m)
	}
	return out
}
