package execlog

import (
	"encoding/json"
	"net/http"
)

// HandleExecLog returns an HTTP handler that exposes the execution log store.
// An optional ?job=<name> query parameter filters results to a single job.
// Without the parameter the full map of job → []Entry is returned.
func HandleExecLog(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		job := r.URL.Query().Get("job")
		if job != "" {
			entries := s.Get(job)
			if entries == nil {
				entries = []Entry{}
			}
			if err := json.NewEncoder(w).Encode(entries); err != nil {
				http.Error(w, "encode error", http.StatusInternalServerError)
			}
			return
		}

		all := s.All()
		if err := json.NewEncoder(w).Encode(all); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
