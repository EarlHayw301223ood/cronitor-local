// Package overlapguard tracks whether a job's previous execution is still
// running when a new tick arrives, and records the number of overlaps skipped.
package overlapguard

import (
	"sync"
	"time"
)

// Guard tracks in-progress executions and detects scheduling overlaps.
type Guard struct {
	mu      sync.Mutex
	running map[string]time.Time // job name -> start time of current execution
	skips   map[string]int       // job name -> cumulative overlap skips
}

// New returns an initialised Guard.
func New() *Guard {
	return &Guard{
		running: make(map[string]time.Time),
		skips:   make(map[string]int),
	}
}

// Enter attempts to mark a job as running. It returns true when the caller
// should proceed, or false when a previous execution is still in progress.
// When false is returned the skip counter for the job is incremented.
func (g *Guard) Enter(job string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, active := g.running[job]; active {
		g.skips[job]++
		return false
	}
	g.running[job] = time.Now()
	return true
}

// Exit marks a job as no longer running. It is a no-op if the job was not
// previously registered via Enter.
func (g *Guard) Exit(job string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.running, job)
}

// Skips returns the number of times a scheduled tick was skipped for job
// because a prior execution was still active.
func (g *Guard) Skips(job string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.skips[job]
}

// Snapshot returns a copy of the current skip counters for all jobs.
func (g *Guard) Snapshot() map[string]int {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make(map[string]int, len(g.skips))
	for k, v := range g.skips {
		out[k] = v
	}
	return out
}

// IsRunning reports whether job currently has an active execution registered.
func (g *Guard) IsRunning(job string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.running[job]
	return ok
}
