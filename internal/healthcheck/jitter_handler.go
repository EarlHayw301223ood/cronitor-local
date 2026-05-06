package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// JitterSource is the subset of jitterdelay.Delayer consumed by HandleJitter.
type JitterSource interface {
	All() map[string]time.Duration
}

// jitterEntry is the JSON representation of a single job's offset.
type jitterEntry struct {
	Job    string  `json:"job"`
	Offset float64 `json:"offset_seconds"`
}

// HandleJitter returns the current jitter offsets for all known jobs.
//
//	GET /jitter          — all jobs
//	GET /jitter?job=name — single job
func HandleJitter(src JitterSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		all := src.All()
		filter := r.URL.Query().Get("job")

		var entries []jitterEntry
		for job, off := range all {
			if filter != "" && job != filter {
				continue
			}
			entries = append(entries, jitterEntry{
				Job:    job,
				Offset: off.Seconds(),
			})
		}
		if entries == nil {
			entries = []jitterEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries) //nolint:errcheck
	}
}
