package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// ThrottleSnapshotter is the subset of throttle.Store used by the handler.
type ThrottleSnapshotter interface {
	Snapshot() map[string]time.Time
}

// throttleEntry is the JSON shape returned per job.
type throttleEntry struct {
	Job         string    `json:"job"`
	LastDispatch time.Time `json:"last_dispatch"`
}

// HandleThrottle returns the last dispatch time for all (or a single) job.
//
// GET /throttle          — all jobs
// GET /throttle?job=name — single job
func HandleThrottle(store ThrottleSnapshotter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		snap := store.Snapshot()
		filter := r.URL.Query().Get("job")

		var entries []throttleEntry
		for job, ts := range snap {
			if filter != "" && job != filter {
				continue
			}
			entries = append(entries, throttleEntry{Job: job, LastDispatch: ts})
		}
		if entries == nil {
			entries = []throttleEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries) //nolint:errcheck
	}
}
