package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/example/cronitor-local/internal/execlog"
)

// execlogStore is the subset of execlog.Store used by the handler.
type execlogStore interface {
	Get(jobName string) []execlog.Entry
	All() map[string][]execlog.Entry
}

// HandleExecLog serves GET /health/execlog[?job=<name>].
//
// Without a query parameter it returns all captured output grouped by job.
// With ?job=<name> it returns only the entries for that job (empty array when
// the job is unknown).
func HandleExecLog(store execlogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		jobName := r.URL.Query().Get("job")
		if jobName != "" {
			entries := store.Get(jobName)
			if entries == nil {
				entries = []execlog.Entry{}
			}
			if err := json.NewEncoder(w).Encode(entries); err != nil {
				http.Error(w, "encode error", http.StatusInternalServerError)
			}
			return
		}

		all := store.All()
		if all == nil {
			all = map[string][]execlog.Entry{}
		}
		if err := json.NewEncoder(w).Encode(all); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
