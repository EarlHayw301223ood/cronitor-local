package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stubAlertStore implements AlertStore for testing.
type stubAlertStore struct {
	records map[string][]AlertRecord
}

func (s *stubAlertStore) All() []AlertRecord {
	var out []AlertRecord
	for _, recs := range s.records {
		out = append(out, recs...)
	}
	return out
}

func (s *stubAlertStore) Get(jobName string) []AlertRecord {
	return s.records[jobName]
}

func makeAlertStore() *stubAlertStore {
	now := time.Now().UTC()
	return &stubAlertStore{
		records: map[string][]AlertRecord{
			"backup": {
				{JobName: "backup", Reason: "failed", FiredAt: now, Message: "exit code 1"},
			},
			"sync": {
				{JobName: "sync", Reason: "overdue", FiredAt: now, Message: "missed window"},
			},
		},
	}
}

func TestHandleAlerts_AllAlerts(t *testing.T) {
	store := makeAlertStore()
	handler := HandleAlerts(store)

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got []AlertRecord
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 alerts, got %d", len(got))
	}
}

func TestHandleAlerts_FilterByJob(t *testing.T) {
	store := makeAlertStore()
	handler := HandleAlerts(store)

	req := httptest.NewRequest(http.MethodGet, "/alerts?job=backup", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got []AlertRecord
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 1 || got[0].JobName != "backup" {
		t.Errorf("expected 1 backup alert, got %+v", got)
	}
}

func TestHandleAlerts_UnknownJob_ReturnsEmptyArray(t *testing.T) {
	store := makeAlertStore()
	handler := HandleAlerts(store)

	req := httptest.NewRequest(http.MethodGet, "/alerts?job=ghost", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got []AlertRecord
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty array, got %+v", got)
	}
}

func TestHandleAlerts_MethodNotAllowed(t *testing.T) {
	store := makeAlertStore()
	handler := HandleAlerts(store)

	req := httptest.NewRequest(http.MethodPost, "/alerts", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
