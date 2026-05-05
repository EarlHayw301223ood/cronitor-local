// Package ratelimit provides per-job alert rate limiting to prevent
// notification storms when a job fails repeatedly.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks the last alert time per job and suppresses duplicate
// alerts within a configurable cooldown window.
type Limiter struct {
	mu       sync.Mutex
	lastSent map[string]time.Time
	cooldown time.Duration
}

// New returns a Limiter with the given cooldown duration.
// Alerts for the same job are suppressed until cooldown has elapsed.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		lastSent: make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Allow reports whether an alert for jobName should be sent.
// It returns true the first time a job is seen, and again only after
// the cooldown window has elapsed since the last allowed alert.
func (l *Limiter) Allow(jobName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	last, seen := l.lastSent[jobName]
	if !seen || time.Since(last) >= l.cooldown {
		l.lastSent[jobName] = time.Now()
		return true
	}
	return false
}

// Reset clears the rate-limit record for jobName, allowing the next
// alert to be sent immediately regardless of the cooldown window.
func (l *Limiter) Reset(jobName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lastSent, jobName)
}

// Snapshot returns a copy of the current last-sent times keyed by job name.
func (l *Limiter) Snapshot() map[string]time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]time.Time, len(l.lastSent))
	for k, v := range l.lastSent {
		out[k] = v
	}
	return out
}
