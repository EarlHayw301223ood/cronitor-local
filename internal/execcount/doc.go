// Package execcount tracks the total number of executions and failures
// for each monitored job. It provides a lightweight counter store that
// accumulates run statistics over the lifetime of the daemon, enabling
// the health endpoint to surface per-job success/failure ratios.
package execcount
