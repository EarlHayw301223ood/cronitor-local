package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/cronitor-local/internal/metrics"
)

// metricsResponse is the JSON shape returned by the /metrics endpoint.
type metricsResponse struct {
	Job         string        `json:"job"`
	TotalRuns   int64         `json:"total_runs"`
	Failures    int64         `json:"failures"`
	LastRun     time.Time     `json:"last_run,omitempty"`
	LastSuccess time.Time     `json:"last_success,omitempty"`
	LastFailure time.Time     `json:"last_failure,omitempty"`
	AvgDuration time.Duration `json:"avg_duration_ms"`
}

// HandleMetrics serves a JSON array of runtime metrics for all tracked jobs.
// It is intended to be registered on the existing health-check HTTP server.
func HandleMetrics(col *metrics.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all := col.All()
		resp := make([]metricsResponse, 0, len(all))
		for _, m := range all {
			resp = append(resp, metricsResponse{
				Job:         m.Name,
				TotalRuns:   m.TotalRuns,
				Failures:    m.Failures,
				LastRun:     m.LastRun,
				LastSuccess: m.LastSuccess,
				LastFailure: m.LastFailure,
				AvgDuration: m.AvgDuration / time.Millisecond,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
		}
	}
}
