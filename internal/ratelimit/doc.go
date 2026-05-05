// Package ratelimit provides per-job alert rate limiting for cronitor-local.
//
// A Limiter suppresses repeated alerts for the same job within a configurable
// cooldown window, preventing notification storms caused by persistently
// failing or overdue jobs.
//
// Usage:
//
//	limiter := ratelimit.New(15 * time.Minute)
//	if limiter.Allow(job.Name) {
//		notifier.Ping(ctx, job.Name, "fail")
//	}
package ratelimit
