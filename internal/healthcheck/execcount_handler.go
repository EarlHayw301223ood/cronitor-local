package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/exampleorg/cronitor-local/internal/execcount"
)

// ExecCountStore is the read interface required by HandleExecCount.
type ExecCountStore interface {
	Get(job string) (execcount.Counter, bool)
	All() map[string]execcount.Counter
}

// HandleExecCount serves execution count statistics for all jobs, or a
// single job when the "job" query parameter is provided.
//
//	GET /health/execcounts
//	GET /health/execcounts?job=backup
func HandleExecCount(store ExecCountStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if job := r.URL.Query().Get("job"); job != "" {
			counter, ok := store.Get(job)
			if !ok {
				// Return a zero counter so callers can distinguish
				// "unknown job" from "no runs yet" via the Total field.
				_ = json.NewEncoder(w).Encode(map[string]execcount.Counter{
					job: {},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]execcount.Counter{
				job: counter,
			})
			return
		}

		_ = json.NewEncoder(w).Encode(store.All())
	}
}
