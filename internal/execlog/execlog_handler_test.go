package execlog_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/execlog"
)

func makeStore(t *testing.T) *execlog.Store {
	t.Helper()
	return execlog.New(10)
}

func seedStore(s *execlog.Store) {
	s.Record(execlog.Entry{
		Job:       "backup",
		StartedAt: time.Now().Add(-2 * time.Minute),
		Duration:  30 * time.Second,
		ExitCode:  0,
		Output:    "backup complete",
		Success:   true,
	})
	s.Record(execlog.Entry{
		Job:       "cleanup",
		StartedAt: time.Now().Add(-1 * time.Minute),
		Duration:  5 * time.Second,
		ExitCode:  1,
		Output:    "error: disk full",
		Success:   false,
	})
}

func TestHandleExecLog_AllEntries(t *testing.T) {
	s := makeStore(t)
	seedStore(s)

	req := httptest.NewRequest(http.MethodGet, "/execlog", nil)
	rw := httptest.NewRecorder()
	execlog.HandleExecLog(s)(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var result map[string][]execlog.Entry
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(result))
	}
}

func TestHandleExecLog_FilterByJob(t *testing.T) {
	s := makeStore(t)
	seedStore(s)

	req := httptest.NewRequest(http.MethodGet, "/execlog?job=backup", nil)
	rw := httptest.NewRecorder()
	execlog.HandleExecLog(s)(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var result []execlog.Entry
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 entry for backup, got %d", len(result))
	}
	if result[0].Job != "backup" {
		t.Errorf("unexpected job: %s", result[0].Job)
	}
}

func TestHandleExecLog_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	s := makeStore(t)
	seedStore(s)

	req := httptest.NewRequest(http.MethodGet, "/execlog?job=ghost", nil)
	rw := httptest.NewRecorder()
	execlog.HandleExecLog(s)(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var result []execlog.Entry
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d entries", len(result))
	}
}

func TestHandleExecLog_MethodNotAllowed(t *testing.T) {
	s := makeStore(t)

	req := httptest.NewRequest(http.MethodPost, "/execlog", nil)
	rw := httptest.NewRecorder()
	execlog.HandleExecLog(s)(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rw.Code)
	}
}
