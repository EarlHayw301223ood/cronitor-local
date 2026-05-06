package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/metrics"
)

func makeMetrics() *metrics.Store {
	s := metrics.New()
	s.RecordRun("backup", true, 2*time.Second)
	s.RecordRun("backup", false, 1*time.Second)
	s.RecordRun("cleanup", true, 500*time.Millisecond)
	return s
}

// decodeMetricsResponse decodes the JSON body from a metrics handler response
// into a map of job name to Snapshot. Fails the test on decode error.
func decodeMetricsResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]metrics.Snapshot {
	t.Helper()
	var result map[string]metrics.Snapshot
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}

func TestHandleMetrics_ReturnsAllJobs(t *testing.T) {
	store := makeMetrics()
	h := &Server{metrics: store}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.HandleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	result := decodeMetricsResponse(t, w)

	if _, ok := result["backup"]; !ok {
		t.Error("expected 'backup' key in response")
	}
	if _, ok := result["cleanup"]; !ok {
		t.Error("expected 'cleanup' key in response")
	}
}

func TestHandleMetrics_CorrectCounts(t *testing.T) {
	store := makeMetrics()
	h := &Server{metrics: store}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.HandleMetrics(w, req)

	result := decodeMetricsResponse(t, w)

	snap := result["backup"]
	if snap.TotalRuns != 2 {
		t.Errorf("expected 2 total runs for backup, got %d", snap.TotalRuns)
	}
	if snap.FailedRuns != 1 {
		t.Errorf("expected 1 failed run for backup, got %d", snap.FailedRuns)
	}
}

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	store := metrics.New()
	h := &Server{metrics: store}

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	w := httptest.NewRecorder()
	h.HandleMetrics(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
