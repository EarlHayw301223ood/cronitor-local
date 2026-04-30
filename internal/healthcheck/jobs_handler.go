package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronitor-local/internal/scheduler"
)

// JobSummary represents a single job's configuration and current status.
type JobSummary struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Status   string `json:"status"`
	LastRun  string `json:"last_run,omitempty"`
}

// HandleJobs returns a JSON list of all registered jobs and their current status.
func (s *Server) HandleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses := s.scheduler.Status()
	summaries := make([]JobSummary, 0, len(statuses))

	for _, st := range statuses {
		lastRun := ""
		if !st.LastRun.IsZero() {
			lastRun = st.LastRun.Format("2006-01-02T15:04:05Z07:00")
		}

		status := "ok"
		if st.LastError != nil {
			status = "error"
		}

		summaries = append(summaries, JobSummary{
			Name:     st.Name,
			Schedule: st.Schedule,
			Command:  st.Command,
			Status:   status,
			LastRun:  lastRun,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summaries)
}

// schedulerStatusProvider abstracts the scheduler for testability.
type schedulerStatusProvider interface {
	Status() []scheduler.JobStatus
}
