package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"
)

// AlertRecord represents a single alert that was fired for a job.
type AlertRecord struct {
	JobName   string    `json:"job_name"`
	Reason    string    `json:"reason"` // "failed" or "overdue"
	FiredAt   time.Time `json:"fired_at"`
	Message   string    `json:"message"`
}

// AlertStore is the read interface required by the alerts handler.
type AlertStore interface {
	All() []AlertRecord
	Get(jobName string) []AlertRecord
}

// HandleAlerts serves GET /alerts and GET /alerts?job=<name>.
func HandleAlerts(store AlertStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var records []AlertRecord
		if job := r.URL.Query().Get("job"); job != "" {
			records = store.Get(job)
		} else {
			records = store.All()
		}

		// Always return a JSON array, never null.
		if records == nil {
			records = []AlertRecord{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}
