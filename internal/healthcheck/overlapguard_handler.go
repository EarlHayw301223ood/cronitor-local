package healthcheck

import (
	"encoding/json"
	"net/http"
)

// OverlapGuardStore is satisfied by *overlapguard.Guard.
type OverlapGuardStore interface {
	Snapshot() map[string]int
	Skips(job string) int
	IsRunning(job string) bool
}

type overlapGuardEntry struct {
	Job     string `json:"job"`
	Skips   int    `json:"skips"`
	Running bool   `json:"running"`
}

// HandleOverlapGuard returns overlap-skip statistics for all jobs, or a single
// job when the "job" query parameter is supplied.
//
//	GET /debug/overlapguard
//	GET /debug/overlapguard?job=backup
func HandleOverlapGuard(store OverlapGuardStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		filter := r.URL.Query().Get("job")
		snap := store.Snapshot()

		var entries []overlapGuardEntry

		if filter != "" {
			entries = []overlapGuardEntry{{
				Job:     filter,
				Skips:   store.Skips(filter),
				Running: store.IsRunning(filter),
			}}
		} else {
			for job, skips := range snap {
				entries = append(entries, overlapGuardEntry{
					Job:     job,
					Skips:   skips,
					Running: store.IsRunning(job),
				})
			}
		}

		if entries == nil {
			entries = []overlapGuardEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries) //nolint:errcheck
	}
}
