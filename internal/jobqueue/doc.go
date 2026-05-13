// Package jobqueue implements a bounded, thread-safe FIFO queue used to hold
// pending job-execution requests when the runner is temporarily unavailable.
//
// When the queue reaches its configured capacity the oldest entry is evicted
// to make room for the incoming one, and the eviction is counted per-job so
// that operators can detect chronic back-pressure through the /queue health
// endpoint.
package jobqueue
