package ratelimit

import (
	"encoding/json"
	"net/http"
)

// Handler returns an HTTP handler that exposes rate-limit state.
func (rl *RateLimit) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		job := r.URL.Query().Get("job")

		rl.mu.Lock()
		defer rl.mu.Unlock()

		type entry struct {
			Job      string `json:"job"`
			LastSent string `json:"last_sent"`
		}

		var results []entry
		for k, v := range rl.records {
			if job != "" && k != job {
				continue
			}
			results = append(results, entry{
				Job:      k,
				LastSent: v.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
		if results == nil {
			results = []entry{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results)
	}
}
