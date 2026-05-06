package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/healthcheck"
	"github.com/user/cronitor-local/internal/throttle"
)

func makeThrottle() *throttle.Throttle {
	t := throttle.New(5 * time.Second)
	t.Allow("job-a")
	t.Allow("job-b")
	return t
}

func TestHandleThrottle_ReturnsAllJobs(t *testing.T) {
	th := makeThrottle()
	h := healthcheck.HandleThrottle(th)

	req := httptest.NewRequest(http.MethodGet, "/throttle", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := result["job-a"]; !ok {
		t.Error("expected job-a in response")
	}
	if _, ok := result["job-b"]; !ok {
		t.Error("expected job-b in response")
	}
}

func TestHandleThrottle_FilterByJob(t *testing.T) {
	th := makeThrottle()
	h := healthcheck.HandleThrottle(th)

	req := httptest.NewRequest(http.MethodGet, "/throttle?job=job-a", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := result["job-a"]; !ok {
		t.Error("expected job-a in filtered response")
	}
	if _, ok := result["job-b"]; ok {
		t.Error("did not expect job-b in filtered response")
	}
}

func TestHandleThrottle_UnknownJob_ReturnsEmptyObject(t *testing.T) {
	th := makeThrottle()
	h := healthcheck.HandleThrottle(th)

	req := httptest.NewRequest(http.MethodGet, "/throttle?job=unknown", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty object, got %v", result)
	}
}

func TestHandleThrottle_MethodNotAllowed(t *testing.T) {
	th := makeThrottle()
	h := healthcheck.HandleThrottle(th)

	req := httptest.NewRequest(http.MethodPost, "/throttle", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
