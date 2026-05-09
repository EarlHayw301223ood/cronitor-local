// Package jitterdelay provides a per-job randomised start-time offset to
// prevent multiple cron jobs from firing simultaneously and overwhelming
// downstream services such as the Cronitor notification endpoint.
//
// A Delayer is safe for concurrent use. Each job receives a stable offset
// for the lifetime of the process; call Reset to force a new value (e.g.
// after a configuration reload).
//
// # Usage
//
// Create a Delayer with a maximum jitter window, then call Wait before
// executing each job:
//
//		d := jitterdelay.New(10 * time.Second)
//		d.Wait("my-job-name") // blocks for a stable random duration ≤ 10s
//
// The same job name always resolves to the same delay within a process
// lifetime, ensuring predictable spacing without repeated randomisation.
package jitterdelay
