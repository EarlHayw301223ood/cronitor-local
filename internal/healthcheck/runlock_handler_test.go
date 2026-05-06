package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/owner/cronitor-local/internal/healthcheck"
	"github.com/owner/cronitor-local/internal/runlock"
)

func makeRunLock(t *testing.T) *runlock.Store {
	t.Helper()
	return runlock.New()
}

func decodeRunLock(t *testing.T, body []byte) []runlock.LockStatus {
	t.Helper()
	var out []runlock.LockStatus
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	return out
}

func TestHandleRunLock_ReturnsAllJobs(t *testing.T) {
	store := makeRunLock(t)
	store.Acquire("backup")
	store.Acquire("report")

	req := httptest.NewRequest(http.MethodGet, "/runlock", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleRunLock(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	entries := decodeRunLock(t, rec.Body.Bytes())
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleRunLock_FilterByJob(t *testing.T) {
	store := makeRunLock(t)
	store.Acquire("sync")

	req := httptest.NewRequest(http.MethodGet, "/runlock?job=sync", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleRunLock(store)(rec, req)

	entries := decodeRunLock(t, rec.Body.Bytes())
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Job != "sync" {
		t.Fatalf("expected job 'sync', got %q", entries[0].Job)
	}
	if !entries[0].Running {
		t.Fatal("expected running to be true")
	}
}

func TestHandleRunLock_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	store := makeRunLock(t)

	req := httptest.NewRequest(http.MethodGet, "/runlock?job=ghost", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleRunLock(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	entries := decodeRunLock(t, rec.Body.Bytes())
	if len(entries) != 0 {
		t.Fatalf("expected empty array, got %d entries", len(entries))
	}
}

func TestHandleRunLock_MethodNotAllowed(t *testing.T) {
	store := makeRunLock(t)

	req := httptest.NewRequest(http.MethodPost, "/runlock", nil)
	rec := httptest.NewRecorder()
	healthcheck.HandleRunLock(store)(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
