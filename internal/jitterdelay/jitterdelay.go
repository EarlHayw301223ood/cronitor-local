// Package jitterdelay provides randomised delay helpers used to spread
// cron job start times and avoid thundering-herd on the notifier endpoint.
package jitterdelay

import (
	"math/rand"
	"sync"
	"time"
)

// Delayer computes a bounded random delay for a named job.
type Delayer struct {
	mu      sync.Mutex
	max     time.Duration
	seed    rand.Source
	rng     *rand.Rand
	offsets map[string]time.Duration
}

// New returns a Delayer whose jitter window is capped at maxJitter.
// Pass 0 to disable jitter entirely.
func New(maxJitter time.Duration) *Delayer {
	src := rand.NewSource(time.Now().UnixNano())
	return &Delayer{
		max:     maxJitter,
		seed:    src,
		rng:     rand.New(src),
		offsets: make(map[string]time.Duration),
	}
}

// For returns a stable per-job jitter offset, computing it once and caching it
// so the same job always gets the same delay within a single process lifetime.
func (d *Delayer) For(job string) time.Duration {
	if d.max == 0 {
		return 0
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if off, ok := d.offsets[job]; ok {
		return off
	}
	off := time.Duration(d.rng.Int63n(int64(d.max)))
	d.offsets[job] = off
	return off
}

// Reset clears the cached offset for job, forcing a new value on the next call
// to For. Useful in tests or after a config reload.
func (d *Delayer) Reset(job string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.offsets, job)
}

// All returns a snapshot of every cached offset keyed by job name.
func (d *Delayer) All() map[string]time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make(map[string]time.Duration, len(d.offsets))
	for k, v := range d.offsets {
		out[k] = v
	}
	return out
}
