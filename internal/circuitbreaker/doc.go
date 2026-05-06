// Package circuitbreaker implements a per-job circuit breaker that prevents
// repeated execution of jobs that have been consistently failing.
//
// The circuit breaker transitions through three states:
//
//   - Closed: normal operation; jobs are allowed to run.
//   - Open: too many consecutive failures have occurred; jobs are blocked
//     until a cooldown period elapses.
//   - Half-Open: the cooldown has elapsed; a single probe execution is
//     permitted to test whether the underlying problem has been resolved.
//     A success closes the breaker; a failure re-opens it.
//
// Each job tracked by the breaker maintains independent state, so a failing
// job does not affect the circuit state of other jobs.
package circuitbreaker
