// Package alertmanager provides alert coordination for cronitor-local.
//
// It inspects job statuses held by the scheduler and fires notifications
// through the notifier when a job has failed (non-zero exit) or has
// missed its expected execution window entirely.
//
// Typical usage:
//
//	mgr := alertmanager.New(n, s, log)
//	// call periodically, e.g. from the watcher tick loop
//	alerts := mgr.CheckAndAlert()
package alertmanager
