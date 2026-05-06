package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/owner/cronitor-local/internal/runlock"
)

// runLockStore is the subset of runlock.Store used by the handler.
type runLockStore interface {
	All() []runlock.LockStatus
	Status(job string) (runlock.LockStatus, bool)
}

// HandleRunLock serves GET /runlock and GET /runlock?job=<name>.
// It exposes which jobs are currently executing and how many concurrent
// executions have been skipped due to an in-flight run.
func HandleRunLock(store runLockStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if job := r.URL.Query().Get("job"); job != "" {
			st, ok := store.Status(job)
			if !ok {
				// Return an empty array for consistency with other handlers.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("[]"))
				return
			}
			json.NewEncoder(w).Encode([]runlock.LockStatus{st})
			return
		}

		all := store.All()
		if all == nil {
			all = []runlock.LockStatus{}
		}
		json.NewEncoder(w).Encode(all)
	}
}
