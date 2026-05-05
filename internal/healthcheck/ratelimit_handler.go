package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// RateLimitStore is the interface the healthcheck server uses to query
// rate-limit state without importing the ratelimit package directly.
type RateLimitStore interface {
	Records() map[string]time.Time
}

// HandleRateLimit returns an HTTP handler that serialises current rate-limit
// records. An optional ?job= query parameter filters results to a single job.
func HandleRateLimit(store RateLimitStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		filter := r.URL.Query().Get("job")
		records := store.Records()

		type entry struct {
			Job      string `json:"job"`
			LastSent string `json:"last_sent"`
		}

		var out []entry
		for job, ts := range records {
			if filter != "" && job != filter {
				continue
			}
			out = append(out, entry{
				Job:      job,
				LastSent: ts.Format(time.RFC3339),
			})
		}
		if out == nil {
			out = []entry{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
