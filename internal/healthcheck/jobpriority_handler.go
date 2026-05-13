package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/cronitor-local/internal/jobpriority"
)

// priorityStore is the subset of jobpriority.Store used by the handler.
type priorityStore interface {
	All() []jobpriority.Entry
	Get(job string) (jobpriority.Level, bool)
}

// HandleJobPriority serves GET /health/priority.
// An optional ?job= query parameter filters results to a single job.
func HandleJobPriority(store priorityStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if job := r.URL.Query().Get("job"); job != "" {
			lvl, ok := store.Get(job)
			if !ok {
				_ = json.NewEncoder(w).Encode([]jobpriority.Entry{})
				return
			}
			_ = json.NewEncoder(w).Encode([]jobpriority.Entry{
				{Job: job, Priority: lvl},
			})
			return
		}

		entries := store.All()
		if entries == nil {
			entries = []jobpriority.Entry{}
		}
		_ = json.NewEncoder(w).Encode(entries)
	}
}
