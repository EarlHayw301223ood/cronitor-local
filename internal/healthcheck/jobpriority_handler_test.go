package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/cronitor-local/internal/healthcheck"
	"github.com/your-org/cronitor-local/internal/jobpriority"
)

func makePriorityStore(t *testing.T) *jobpriority.Store {
	t.Helper()
	return jobpriority.New()
}

func TestHandleJobPriority_ReturnsAllJobs(t *testing.T) {
	s := makePriorityStore(t)
	_ = s.Set("backup", jobpriority.High)
	_ = s.Set("report", jobpriority.Low)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/priority", nil)
	healthcheck.HandleJobPriority(s)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var entries []jobpriority.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
}

func TestHandleJobPriority_FilterByJob(t *testing.T) {
	s := makePriorityStore(t)
	_ = s.Set("backup", jobpriority.High)
	_ = s.Set("report", jobpriority.Normal)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/priority?job=backup", nil)
	healthcheck.HandleJobPriority(s)(rec, req)

	var entries []jobpriority.Entry
	_ = json.NewDecoder(rec.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Priority != jobpriority.High {
		t.Fatalf("want High, got %d", entries[0].Priority)
	}
}

func TestHandleJobPriority_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	s := makePriorityStore(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/priority?job=ghost", nil)
	healthcheck.HandleJobPriority(s)(rec, req)

	var entries []jobpriority.Entry
	_ = json.NewDecoder(rec.Body).Decode(&entries)
	if len(entries) != 0 {
		t.Fatalf("want empty array, got %d entries", len(entries))
	}
}

func TestHandleJobPriority_MethodNotAllowed(t *testing.T) {
	s := makePriorityStore(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health/priority", nil)
	healthcheck.HandleJobPriority(s)(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}
