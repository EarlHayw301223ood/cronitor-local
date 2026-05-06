// Package throttle enforces a minimum dispatch interval for each monitored
// job. It prevents burst re-runs that can occur when the scheduler tick
// interval is shorter than a job's natural cadence, or when clock skew
// causes multiple ticks to fire in rapid succession.
//
// Usage:
//
//	store := throttle.New(30 * time.Second)
//	if store.Allow(jobName) {
//	    // dispatch the job
//	}
package throttle
