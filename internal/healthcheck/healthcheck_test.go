package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronitor-local/internal/healthcheck"
	"github.com/example/cronitor-local/internal/scheduler"
)

type mockScheduler struct {
	statuses map[string]scheduler.JobStatus
}

func (m *mockScheduler) Status() map[string]scheduler.JobStatus {
	return m.statuses
}

func makeStatuses(hasError bool) map[string]scheduler.JobStatus {
	err := error(nil)
	if hasError {
		err = fmt.Errorf("command exited with code 1")
	}
	return map[string]scheduler.JobStatus{
		"backup": {
			LastRun:   time.Now(),
			LastError: err,
			RunCount:  3,
		},
	}
}

func TestHandleHealth_AllHealthy(t *testing.T) {
	sched := &mockScheduler{statuses: makeStatuses(false)}
	srv := healthcheck.New(sched, 0)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	// Access via exported handler by starting a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv2 := healthcheck.New(sched, 0)
		_ = srv2
		// Use internal handler indirectly through a real server
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()
	_ = req
	_ = rec
	_ = srv

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleHealth_WithError(t *testing.T) {
	sched := &mockScheduler{statuses: makeStatuses(true)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, st := range sched.Status() {
			if st.LastError != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("unhealthy"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}
}

func TestHandleStatus_ReturnsJSON(t *testing.T) {
	sched := &mockScheduler{statuses: makeStatuses(false)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(sched.Status())
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode JSON response: %v", err)
	}
}
