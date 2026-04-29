// Package watcher ties together the scheduler, runner, and notifier to
// orchestrate periodic job execution.
//
// The Watcher polls the Scheduler on a configurable interval. For every job
// that is due at the current minute it launches a goroutine that:
//  1. Sends a "run" ping via the Notifier.
//  2. Executes the command via the Runner.
//  3. Sends either a "complete" or "fail" ping depending on the outcome.
//  4. Records the result back in the Scheduler for status reporting.
//
// The watch loop respects context cancellation, making it straightforward to
// integrate with OS signal handling in main.
package watcher
