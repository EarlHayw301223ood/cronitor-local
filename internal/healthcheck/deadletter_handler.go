package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/cronitor-local/internal/deadletter"
)

// DeadLetterStore is the read interface required by HandleDeadLetter.
type DeadLetterStore interface {
	All() []deadletter.Entry
	ForJob(name string) []deadletter.Entry
}

// HandleDeadLetter returns dead-letter entries as JSON.
// Optional query param: ?job=<name> to filter by job.
func HandleDeadLetter(store DeadLetterStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var entries []deadletter.Entry
		if job := r.URL.Query().Get("job"); job != "" {
			entries = store.ForJob(job)
		} else {
			entries = store.All()
		}

		// Always encode an array, never null.
		if entries == nil {
			entries = []deadletter.Entry{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
