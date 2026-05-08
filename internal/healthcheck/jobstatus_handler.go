package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/example/cronitor-local/internal/jobstatus"
)

// JobStatusStore is the read interface required by HandleJobStatus.
type JobStatusStore interface {
	All() []jobstatus.Entry
	Get(job string) (jobstatus.Entry, bool)
}

// HandleJobStatus serves the current status of all jobs, or a single job when
// the "job" query parameter is provided.
//
// GET /status          → all jobs
// GET /status?job=name → single job (404 if unknown)
func HandleJobStatus(store JobStatusStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if name := r.URL.Query().Get("job"); name != "" {
			e, ok := store.Get(name)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "job not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(e)
			return
		}

		_ = json.NewEncoder(w).Encode(store.All())
	}
}
