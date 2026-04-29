package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronitor-local/internal/notifier"
)

func TestPing_Success(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.New("test-api-key")
	n.BaseURL = ts.URL

	err := n.Ping(notifier.Event{
		JobName:  "backup",
		Status:   "complete",
		Duration: 1.5,
		Message:  "ok",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received["status"] != "complete" {
		t.Errorf("expected status=complete, got %v", received["status"])
	}
	if received["duration"].(float64) != 1.5 {
		t.Errorf("expected duration=1.5, got %v", received["duration"])
	}
}

func TestPing_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notifier.New("test-api-key")
	n.BaseURL = ts.URL

	err := n.Ping(notifier.Event{JobName: "backup", Status: "fail"})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestPing_NetworkError(t *testing.T) {
	n := notifier.New("test-api-key")
	n.BaseURL = "http://127.0.0.1:1" // nothing listening

	err := n.Ping(notifier.Event{JobName: "backup", Status: "run"})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}

func TestPing_URLContainsJobName(t *testing.T) {
	var capturedPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.New("my-key")
	n.BaseURL = ts.URL

	_ = n.Ping(notifier.Event{JobName: "daily-report", Status: "run"})

	if capturedPath != "/my-key/daily-report" {
		t.Errorf("unexpected path: %s", capturedPath)
	}
}
