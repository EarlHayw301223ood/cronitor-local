package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/execlog"
	"github.com/your-org/cronitor-local/internal/healthcheck"
)

func makeExecLogStore(t *testing.T) *execlog.Store {
	t.Helper()
	s := execlog.New(20)
	s.Record(execlog.Entry{
		Job:       "report",
		StartedAt: time.Now().Add(-10 * time.Minute),
		Duration:  15 * time.Second,
		ExitCode:  0,
		Output:    "report sent",
		Success:   true,
	})
	s.Record(execlog.Entry{
		Job:       "report",
		StartedAt: time.Now().Add(-5 * time.Minute),
		Duration:  12 * time.Second,
		ExitCode:  0,
		Output:    "report sent again",
		Success:   true,
	})
	return s
}

func TestHealthcheck_ExecLogRoute_ReturnsOK(t *testing.T) {
	store := makeExecLogStore(t)

	req := httptest.NewRequest(http.MethodGet, "/execlog", nil)
	rw := httptest.NewRecorder()

	healthcheck.HandleExecLog(store)(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	ct := rw.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestHealthcheck_ExecLogRoute_FilterByJob(t *testing.T) {
	store := makeExecLogStore(t)

	req := httptest.NewRequest(http.MethodGet, "/execlog?job=report", nil)
	rw := httptest.NewRecorder()

	healthcheck.HandleExecLog(store)(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var entries []execlog.Entry
	if err := json.NewDecoder(rw.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for report, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Job != "report" {
			t.Errorf("unexpected job in result: %s", e.Job)
		}
	}
}

func TestHealthcheck_ExecLogRoute_MethodNotAllowed(t *testing.T) {
	store := makeExecLogStore(t)

	req := httptest.NewRequest(http.MethodDelete, "/execlog", nil)
	rw := httptest.NewRecorder()

	healthcheck.HandleExecLog(store)(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rw.Code)
	}
}
