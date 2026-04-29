// Package history provides a thread-safe, per-job ring-buffer of execution
// history entries. Each entry captures the start time, duration, success
// flag, and captured output of a single job run.
//
// Usage:
//
//	h := history.New(100) // keep last 100 entries per job
//	h.Record(history.Entry{JobName: "backup", Success: true, ...})
//	entries := h.Get("backup")
package history
