package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubJitter struct {
	data map[string]time.Duration
}

func (s *stubJitter) All() map[string]time.Duration { return s.data }

func makeJitter() *stubJitter {
	return &stubJitter{
		data: map[string]time.Duration{
			"backup": 3 * time.Second,
			"report": 7500 * time.Millisecond,
		},
	}
}

func TestHandleJitter_ReturnsAllJobs(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jitter", nil)
	HandleJitter(makeJitter())(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []jitterEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleJitter_FilterByJob(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jitter?job=backup", nil)
	HandleJitter(makeJitter())(rec, req)

	var entries []jitterEntry
	json.NewDecoder(rec.Body).Decode(&entries) //nolint:errcheck
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Job != "backup" {
		t.Errorf("expected job 'backup', got %q", entries[0].Job)
	}
	if entries[0].Offset != 3.0 {
		t.Errorf("expected offset 3.0, got %v", entries[0].Offset)
	}
}

func TestHandleJitter_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jitter?job=ghost", nil)
	HandleJitter(makeJitter())(rec, req)

	var entries []jitterEntry
	json.NewDecoder(rec.Body).Decode(&entries) //nolint:errcheck
	if len(entries) != 0 {
		t.Fatalf("expected empty array, got %d entries", len(entries))
	}
}

func TestHandleJitter_MethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jitter", nil)
	HandleJitter(makeJitter())(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
