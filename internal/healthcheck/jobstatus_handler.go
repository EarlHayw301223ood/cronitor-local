package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/cronitor-local/internal/jobstatus"
)

// HandleJobStatus returns an HTTP handler that reports the last known
// execution status for each monitored job.
//
// Query parameters:
//
//	job — optional; when provided, only the named job is returned.
//
// Response shape:
//
//	{
//	  "<job-name>": { "ok": true, "at": "<RFC3339>" },
//	  ...
//	}
func HandleJobStatus(store *jobstatus.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		all := store.All()
		result := make(map[string]jobstatus.Entry)

		if job := r.URL.Query().Get("job"); job != "" {
			if entry, ok := store.Get(job); ok {
				result[job] = entry
			}
		} else {
			for k, v := range all {
				result[k] = v
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
