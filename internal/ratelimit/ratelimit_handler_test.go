package ratelimit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/ratelimit"
)

func makeRateLimit() *ratelimit.RateLimit {
	return ratelimit.New(5 * time.Minute)
}

func TestHandleRateLimit_ReturnsJSON(t *testing.T) {
	rl := makeRateLimit()
	rl.Allow("job-a") // record a hit so the map is non-empty

	req := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	rw := httptest.NewRecorder()

	rl.Handler()(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var entries []map[string]string
	if err := json.NewDecoder(rw.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0]["job"] != "job-a" {
		t.Errorf("expected job-a, got %s", entries[0]["job"])
	}
}

func TestHandleRateLimit_MethodNotAllowed(t *testing.T) {
	rl := makeRateLimit()
	req := httptest.NewRequest(http.MethodPost, "/ratelimit", nil)
	rw := httptest.NewRecorder()

	rl.Handler()(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestHandleRateLimit_FilterByJob(t *testing.T) {
	rl := makeRateLimit()
	rl.Allow("job-a")
	rl.Allow("job-b")

	req := httptest.NewRequest(http.MethodGet, "/ratelimit?job=job-a", nil)
	rw := httptest.NewRecorder()

	rl.Handler()(rw, req)

	var entries []map[string]string
	if err := json.NewDecoder(rw.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after filter, got %d", len(entries))
	}
	if entries[0]["job"] != "job-a" {
		t.Errorf("unexpected job: %s", entries[0]["job"])
	}
}
