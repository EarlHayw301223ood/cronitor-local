package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/healthcheck"
	"github.com/user/cronitor-local/internal/ratelimit"
)

func makeRateLimit() *ratelimit.Limiter {
	return ratelimit.New(30 * time.Second)
}

func TestHandleRateLimit_ReturnsJSON(t *testing.T) {
	rl := makeRateLimit()
	rl.Allow("job-a")
	rl.Allow("job-b")

	h := healthcheck.HandleRateLimit(rl)
	req := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := body["entries"]; !ok {
		t.Error("expected 'entries' key in response")
	}
}

func TestHandleRateLimit_MethodNotAllowed(t *testing.T) {
	rl := makeRateLimit()
	h := healthcheck.HandleRateLimit(rl)
	req := httptest.NewRequest(http.MethodPost, "/ratelimit", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleRateLimit_FilterByJob(t *testing.T) {
	rl := makeRateLimit()
	rl.Allow("job-x")
	rl.Allow("job-y")

	h := healthcheck.HandleRateLimit(rl)
	req := httptest.NewRequest(http.MethodGet, "/ratelimit?job=job-x", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	entries, ok := body["entries"].([]interface{})
	if !ok {
		t.Fatal("entries is not an array")
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry for job-x, got %d", len(entries))
	}
}
