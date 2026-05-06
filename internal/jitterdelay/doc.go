// Package jitterdelay provides a per-job randomised start-time offset to
// prevent multiple cron jobs from firing simultaneously and overwhelming
// downstream services such as the Cronitor notification endpoint.
//
// A Delayer is safe for concurrent use. Each job receives a stable offset
// for the lifetime of the process; call Reset to force a new value (e.g.
// after a configuration reload).
package jitterdelay
