package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/deadletter"
	"github.com/your-org/cronitor-local/internal/healthcheck"
)

func makeDeadLetterStore(entries ...deadletter.Entry) *deadletter.Store {
	s := deadletter.New(100)
	for _, e := range entries {
		s.Add(e)
	}
	return s
}

func dlEntry(job string) deadletter.Entry {
	return deadletter.Entry{
		JobName:   job,
		Command:   job + ".sh",
		Error:     "exit status 1",
		Attempts:  3,
		CreatedAt: time.Now(),
	}
}

func TestHandleDeadLetter_AllEntries(t *testing.T) {
	store := makeDeadLetterStore(dlEntry("alpha"), dlEntry("beta"))
	h := healthcheck.HandleDeadLetter(store)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/deadletter", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []deadletter.Entry
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

func TestHandleDeadLetter_FilterByJob(t *testing.T) {
	store := makeDeadLetterStore(dlEntry("alpha"), dlEntry("beta"), dlEntry("alpha"))
	h := healthcheck.HandleDeadLetter(store)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/deadletter?job=alpha", nil))

	var result []deadletter.Entry
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries for alpha, got %d", len(result))
	}
}

func TestHandleDeadLetter_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	store := makeDeadLetterStore(dlEntry("alpha"))
	h := healthcheck.HandleDeadLetter(store)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/deadletter?job=ghost", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []deadletter.Entry
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d entries", len(result))
	}
}

func TestHandleDeadLetter_MethodNotAllowed(t *testing.T) {
	store := makeDeadLetterStore()
	h := healthcheck.HandleDeadLetter(store)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/deadletter", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
