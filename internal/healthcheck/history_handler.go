package healthcheck

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/your-org/cronitor-local/internal/history"
)

// historyProvider is the subset of history.History used by the handler.
type historyProvider interface {
	Get(jobName string) []history.Entry
	All() map[string][]history.Entry
}

// HandleHistory serves execution history over HTTP.
//
// GET /history         → all jobs
// GET /history/{name}  → single job
func HandleHistory(h historyProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Trim leading slash and the "history" prefix.
		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.SplitN(path, "/", 2)

		var payload any
		if len(parts) == 2 && parts[1] != "" {
			payload = h.Get(parts[1])
		} else {
			payload = h.All()
		}

		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
