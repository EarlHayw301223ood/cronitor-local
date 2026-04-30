package healthcheck_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/healthcheck"
	"github.com/user/cronitor-local/internal/scheduler"
)

type fakeScheduler struct {
	statuses []scheduler.JobStatus
}

func (f *fakeScheduler) Status() []scheduler.JobStatus {
	return f.statuses
}

func makeJobsScheduler(statuses []scheduler.JobStatus) *fakeScheduler {
	return &fakeScheduler{statuses: statuses}
}

func TestHandleJobs_ReturnsAllJobs(t *testing.T) {
	now := time.Now()
	sched := makeJobsScheduler([]scheduler.JobStatus{
		{Name: "backup", Schedule: "@daily", Command: "/bin/backup.sh", LastRun: now},
		{Name: "cleanup", Schedule: "0 * * * *", Command: "/bin/cleanup.sh", LastRun: now},
	})

	srv := healthcheck.NewWithScheduler(sched)
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()

	srv.HandleJobs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var jobs []healthcheck.JobSummary
	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].Name != "backup" {
		t.Errorf("expected first job to be 'backup', got %q", jobs[0].Name)
	}
}

func TestHandleJobs_ErrorStatus(t *testing.T) {
	sched := makeJobsScheduler([]scheduler.JobStatus{
		{Name: "failing", Schedule: "* * * * *", Command: "/bin/fail.sh",
			LastRun: time.Now(), LastError: fmt.Errorf("exit status 1")},
	})

	srv := healthcheck.NewWithScheduler(sched)
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()

	srv.HandleJobs(rec, req)

	var jobs []healthcheck.JobSummary
	_ = json.NewDecoder(rec.Body).Decode(&jobs)

	if jobs[0].Status != "error" {
		t.Errorf("expected status 'error', got %q", jobs[0].Status)
	}
}

func TestHandleJobs_MethodNotAllowed(t *testing.T) {
	srv := healthcheck.NewWithScheduler(makeJobsScheduler(nil))
	req := httptest.NewRequest(http.MethodPost, "/jobs", nil)
	rec := httptest.NewRecorder()

	srv.HandleJobs(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleJobs_EmptyLastRun(t *testing.T) {
	sched := makeJobsScheduler([]scheduler.JobStatus{
		{Name: "new-job", Schedule: "@hourly", Command: "/bin/new.sh"},
	})

	srv := healthcheck.NewWithScheduler(sched)
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()

	srv.HandleJobs(rec, req)

	var jobs []healthcheck.JobSummary
	_ = json.NewDecoder(rec.Body).Decode(&jobs)

	if jobs[0].LastRun != "" {
		t.Errorf("expected empty last_run for job that has never run, got %q", jobs[0].LastRun)
	}
}
