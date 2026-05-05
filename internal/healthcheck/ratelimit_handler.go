package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// RateLimitSnapshotter is satisfied by ratelimit.Limiter.
type RateLimitSnapshotter interface {
	Snapshot() map[string]time.Time
}

// rateLimitEntry is the JSON shape for a single job's rate-limit record.
type rateLimitEntry struct {
	Job      string    `json:"job"`
	LastSent time.Time `json:"last_sent"`
	Cooldown string    `json:"cooldown"`
}

// HandleRateLimit returns an HTTP handler that exposes the current
// rate-limit snapshot so operators can see which jobs are suppressed.
func HandleRateLimit(snap RateLimitSnapshotter, cooldown time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		state := snap.Snapshot()
		entries := make([]rateLimitEntry, 0, len(state))
		for job, ts := range state {
			entries = append(entries, rateLimitEntry{
				Job:      job,
				LastSent: ts,
				Cooldown: cooldown.String(),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
