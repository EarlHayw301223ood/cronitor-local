package healthcheck

import (
	"encoding/json"
	"net/http"
)

// jobEnvReader is satisfied by jobenv.Store.
type jobEnvReader interface {
	Get(job string) (map[string]string, bool)
	All() map[string]map[string]string
}

// HandleJobEnv serves per-job environment variable overrides.
//
// GET /jobs/env          — returns all jobs' env maps
// GET /jobs/env?job=name — returns the env map for a single job
func HandleJobEnv(store jobEnvReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if job := r.URL.Query().Get("job"); job != "" {
			env, ok := store.Get(job)
			if !ok {
				env = map[string]string{}
			}
			_ = json.NewEncoder(w).Encode(map[string]map[string]string{job: env})
			return
		}

		_ = json.NewEncoder(w).Encode(store.All())
	}
}
