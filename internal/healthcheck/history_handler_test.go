package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/healthcheck"
	"github.com/your-org/cronitor-local/internal/history"
)

func makeHistory() *history.History {
	h := history.New(10)
	h.Record(history.Entry{
		JobName:   "backup",
		StartedAt: time.Now(),
		Duration:  200 * time.Millisecond,
		Success:   true,
		Output:    "ok",
	})
	h.Record(history.Entry{
		JobName:   "cleanup",
		StartedAt: time.Now(),
		Duration:  50 * time.Millisecond,
		Success:   false,
		Output:    "err",
	})
	return h
}

func TestHandleHistory_AllJobs(t *testing.T) {
	h := makeHistory()
	handler := healthcheck.HandleHistory(h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string][]history.Entry
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(result))
	}
}

func TestHandleHistory_SingleJob(t *testing.T) {
	h := makeHistory()
	handler := healthcheck.HandleHistory(h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history/backup", nil)
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []history.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry for 'backup', got %d", len(entries))
	}
	if !entries[0].Success {
		t.Error("expected success=true for backup entry")
	}
}

func TestHandleHistory_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	h := makeHistory()
	handler := healthcheck.HandleHistory(h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history/nonexistent", nil)
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []history.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty array for unknown job, got %d entries", len(entries))
	}
}
