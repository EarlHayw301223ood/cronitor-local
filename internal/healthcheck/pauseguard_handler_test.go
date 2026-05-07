package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/cronitor-local/internal/pauseguard"
)

func makePauseGuard() *pauseguard.Store {
	return pauseguard.New()
}

func TestHandlePauseGuard_ReturnsAllJobs(t *testing.T) {
	store := makePauseGuard()
	store.Pause("alpha", 0)
	store.Pause("beta", 0)

	h := HandlePauseGuardTyped(store)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/pauseguard", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestHandlePauseGuard_FilterByJob(t *testing.T) {
	store := makePauseGuard()
	store.Pause("alpha", 0)
	store.Pause("beta", 0)

	h := HandlePauseGuardTyped(store)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/pauseguard?job=alpha", nil))

	var result []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry for job=alpha, got %d", len(result))
	}
	if result[0]["job"] != "alpha" {
		t.Errorf("expected job=alpha, got %v", result[0]["job"])
	}
}

func TestHandlePauseGuard_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	store := makePauseGuard()
	store.Pause("alpha", 0)

	h := HandlePauseGuardTyped(store)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/pauseguard?job=ghost", nil))

	var result []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Fatalf("expected empty array for unknown job, got %d", len(result))
	}
}

func TestHandlePauseGuard_MethodNotAllowed(t *testing.T) {
	store := makePauseGuard()
	h := HandlePauseGuardTyped(store)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/pauseguard", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
