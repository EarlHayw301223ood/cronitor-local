package healthcheck

import (
	"encoding/json"
	"net/http"
)

// dependencyQuerier is satisfied by *jobdependency.Store.
type dependencyQuerier interface {
	All() map[string][]string
	Ready(job string) (bool, string, error)
}

type dependencyEntry struct {
	Upstreams []string `json:"upstreams"`
	Ready     bool     `json:"ready"`
	Blocking  string   `json:"blocking,omitempty"`
}

// HandleJobDependency serves GET /health/dependencies[?job=name].
// It reports declared upstream dependencies and whether each job is currently
// unblocked.
func HandleJobDependency(store dependencyQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		all := store.All()
		filter := r.URL.Query().Get("job")

		result := make(map[string]dependencyEntry)

		for job, upstreams := range all {
			if filter != "" && job != filter {
				continue
			}
			ready, blocking, _ := store.Ready(job)
			result[job] = dependencyEntry{
				Upstreams: upstreams,
				Ready:     ready,
				Blocking:  blocking,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
