// Package retrypolicy implements a thread-safe, per-job retry budget with
// exponential backoff.
//
// Usage:
//
//	policy := retrypolicy.New(3, 5*time.Second)
//
//	// After a job failure:
//	if policy.ShouldRetry(job.Name) {
//		time.Sleep(policy.Backoff(job.Name))
//		// re-run the job …
//	} else {
//		// dispatch alert, budget exhausted
//	}
//
//	// After a successful run:
//	policy.Reset(job.Name)
package retrypolicy
