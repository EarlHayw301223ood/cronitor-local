package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/healthcheck"
	"github.com/your-org/cronitor-local/internal/jobstatus"
)

func makeJobStatusStore(t *testing.T) *jobstatus.Store {
	t.Helper()
	return jobstatus.New()
}

func TestHandleJobStatus_AllJobs(t *testing.T) {
	store := makeJobStatusStore(t)
	now := time.Now()
	store.Set("backup", jobstatus.Entry{OK: true, At: now})
	store.Set("sync", jobstatus.Entry{OK: false, At: now})

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleJobStatus(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(result))
	}
}

func TestHandleJobStatus_FilterByJob(t *testing.T) {
	store := makeJobStatusStore(t)
	now := time.Now()
	store.Set("backup", jobstatus.Entry{OK: true, At: now})
	store.Set("sync", jobstatus.Entry{OK: false, At: now})

	req := httptest.NewRequest(http.MethodGet, "/status?job=backup", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleJobStatus(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := result["backup"]; !ok {
		t.Error("expected 'backup' key in response")
	}
	if _, ok := result["sync"]; ok {
		t.Error("did not expect 'sync' key in filtered response")
	}
}

func TestHandleJobStatus_UnknownJob_ReturnsEmptyObject(t *testing.T) {
	store := makeJobStatusStore(t)

	req := httptest.NewRequest(http.MethodGet, "/status?job=ghost", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleJobStatus(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty object, got %d entries", len(result))
	}
}

func TestHandleJobStatus_MethodNotAllowed(t *testing.T) {
	store := makeJobStatusStore(t)

	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleJobStatus(store)(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
