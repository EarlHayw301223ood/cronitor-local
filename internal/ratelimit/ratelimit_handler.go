package ratelimit

import "net/http"

// Handler returns an HTTP middleware that enforces rate limiting on incoming
// requests. If the rate limit is exceeded, it responds with HTTP 429 Too Many
// Requests and does not call the next handler.
func Handler(limiter *Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
