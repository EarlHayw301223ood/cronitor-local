// Package circuitbreaker provides a per-job circuit breaker that temporarily
// disables alerting or execution for jobs that repeatedly fail, preventing
// alert storms and reducing noise during sustained outages.
package circuitbreaker

import (
	"sync"
	"time"
)

// State represents the current state of a circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // tripped; calls are rejected
	StateHalfOpen              // probe allowed to test recovery
)

// Breaker is a per-job circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	threshold    int
	cooldown     time.Duration
	counts       map[string]int
	states       map[string]State
	tripTimes    map[string]time.Time
	now          func() time.Time
}

// New returns a Breaker that opens after threshold consecutive failures and
// resets after cooldown has elapsed.
func New(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{
		threshold: threshold,
		cooldown:  cooldown,
		counts:    make(map[string]int),
		states:    make(map[string]State),
		tripTimes: make(map[string]time.Time),
		now:       time.Now,
	}
}

// Allow reports whether the job should be allowed to proceed.
// It transitions StateOpen → StateHalfOpen once the cooldown expires.
func (b *Breaker) Allow(job string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.states[job] {
	case StateOpen:
		if b.now().Sub(b.tripTimes[job]) >= b.cooldown {
			b.states[job] = StateHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

// RecordSuccess resets the failure count and closes the circuit for job.
func (b *Breaker) RecordSuccess(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts[job] = 0
	b.states[job] = StateClosed
}

// RecordFailure increments the failure count for job and opens the circuit
// once the threshold is reached.
func (b *Breaker) RecordFailure(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts[job]++
	if b.counts[job] >= b.threshold && b.states[job] != StateOpen {
		b.states[job] = StateOpen
		b.tripTimes[job] = b.now()
	}
}

// StateOf returns the current State for job.
func (b *Breaker) StateOf(job string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.states[job]
}
