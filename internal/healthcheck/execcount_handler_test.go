package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exampleorg/cronitor-local/internal/execcount"
	"github.com/exampleorg/cronitor-local/internal/healthcheck"
)

// stubExecCountStore satisfies healthcheck.ExecCountStore for testing.
type stubExecCountStore struct {
	data map[string]execcount.Counter
}

func (s *stubExecCountStore) Get(job string) (execcount.Counter, bool) {
	c, ok := s.data[job]
	return c, ok
}

func (s *stubExecCountStore) All() map[string]execcount.Counter {
	return s.data
}

func makeExecCountStore() *stubExecCountStore {
	return &stubExecCountStore{
		data: map[string]execcount.Counter{
			"backup": {Total: 10, Failures: 2},
			"cleanup": {Total: 5, Failures: 0},
		},
	}
}

func TestHandleExecCount_AllJobs(t *testing.T) {
	h := healthcheck.HandleExecCount(makeExecCountStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/execcounts", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]execcount.Counter
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(result))
	}
}

func TestHandleExecCount_FilterByJob(t *testing.T) {
	h := healthcheck.HandleExecCount(makeExecCountStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/execcounts?job=backup", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]execcount.Counter
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	c, ok := result["backup"]
	if !ok {
		t.Fatal("expected 'backup' key in response")
	}
	if c.Total != 10 || c.Failures != 2 {
		t.Errorf("unexpected counter: %+v", c)
	}
}

func TestHandleExecCount_UnknownJob_ReturnsZeroCounter(t *testing.T) {
	h := healthcheck.HandleExecCount(makeExecCountStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/execcounts?job=unknown", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]execcount.Counter
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	c := result["unknown"]
	if c.Total != 0 || c.Failures != 0 {
		t.Errorf("expected zero counter for unknown job, got %+v", c)
	}
}

func TestHandleExecCount_MethodNotAllowed(t *testing.T) {
	h := healthcheck.HandleExecCount(makeExecCountStore())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/health/execcounts", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
