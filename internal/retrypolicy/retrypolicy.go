// Package retrypolicy provides configurable retry logic for job executions.
// When a job fails, the policy determines whether and how many times it should
// be retried before an alert is dispatched.
package retrypolicy

import (
	"sync"
	"time"
)

// Policy holds retry configuration and per-job attempt counters.
type Policy struct {
	mu       sync.Mutex
	attempts map[string]int
	maxRetry int
	backoff  time.Duration
}

// New creates a Policy with the given maximum retry count and base backoff
// duration between attempts.
func New(maxRetry int, backoff time.Duration) *Policy {
	return &Policy{
		attempts: make(map[string]int),
		maxRetry: maxRetry,
		backoff:  backoff,
	}
}

// ShouldRetry returns true when the job has not yet exhausted its retry budget
// and increments the internal attempt counter.
func (p *Policy) ShouldRetry(job string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.attempts[job] < p.maxRetry {
		p.attempts[job]++
		return true
	}
	return false
}

// Attempts returns the current retry attempt count for a job.
func (p *Policy) Attempts(job string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.attempts[job]
}

// Reset clears the attempt counter for a job, typically called after a
// successful execution.
func (p *Policy) Reset(job string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.attempts, job)
}

// Backoff returns the delay that should be observed before the next attempt.
// Each successive attempt doubles the base backoff (exponential).
func (p *Policy) Backoff(job string) time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	attempt := p.attempts[job]
	if attempt <= 0 {
		return p.backoff
	}
	shift := attempt - 1
	if shift > 10 {
		shift = 10 // cap at ~1024× base to avoid overflow
	}
	return p.backoff * (1 << uint(shift))
}
