package healthcheck

import (
	"encoding/json"
	"net/http"
)

// JobQueuer is the subset of jobqueue.Queue used by HandleJobQueue.
type JobQueuer interface {
	Len() int
	Evictions(jobName string) int
	AllEvictions() map[string]int
}

type jobQueueResponse struct {
	Depth     int            `json:"depth"`
	Evictions map[string]int `json:"evictions"`
}

// HandleJobQueue returns the current queue depth and per-job eviction counts.
//
//	GET /queue          – full snapshot
//	GET /queue?job=name – evictions for a single job
func HandleJobQueue(q JobQueuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		jobName := r.URL.Query().Get("job")
		var resp jobQueueResponse

		if jobName != "" {
			resp = jobQueueResponse{
				Depth:     q.Len(),
				Evictions: map[string]int{jobName: q.Evictions(jobName)},
			}
		} else {
			resp = jobQueueResponse{
				Depth:     q.Len(),
				Evictions: q.AllEvictions(),
			}
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
