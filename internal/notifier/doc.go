// Package notifier provides a client for sending job lifecycle events
// (run, complete, fail) to the Cronitor monitoring API.
//
// Usage:
//
//	n := notifier.New("your-api-key")
//	err := n.Ping(notifier.Event{
//		JobName: "nightly-backup",
//		Status:  "complete",
//		Duration: 42.3,
//	})
//
// The BaseURL field can be overridden for testing or self-hosted deployments.
package notifier
