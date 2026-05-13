package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/cronitor-local/internal/healthcheck"
	"github.com/yourorg/cronitor-local/internal/jobqueue"
)

func makeQueue() *jobqueue.Queue {
	return jobqueue.New(8)
}

func TestHandleJobQueue_ReturnsDepthAndEvictions(t *testing.T) {
	q := makeQueue()
	q.Enqueue("backup")
	q.Enqueue("report")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/queue", nil)
	healthcheck.HandleJobQueue(q)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Depth     int            `json:"depth"`
		Evictions map[string]int `json:"evictions"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body.Depth != 2 {
		t.Fatalf("expected depth 2, got %d", body.Depth)
	}
}

func TestHandleJobQueue_FilterByJob_ReturnsOnlyThatJob(t *testing.T) {
	q := jobqueue.New(1)
	q.Enqueue("victim")
	q.Enqueue("new") // evicts victim

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/queue?job=victim", nil)
	healthcheck.HandleJobQueue(q)(rec, req)

	var body struct {
		Evictions map[string]int `json:"evictions"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Evictions["victim"] != 1 {
		t.Fatalf("expected 1 eviction for victim, got %d", body.Evictions["victim"])
	}
	if _, ok := body.Evictions["new"]; ok {
		t.Fatal("unexpected key 'new' in filtered response")
	}
}

func TestHandleJobQueue_MethodNotAllowed(t *testing.T) {
	q := makeQueue()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/queue", nil)
	healthcheck.HandleJobQueue(q)(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleJobQueue_EmptyQueue_ReturnsZeroDepth(t *testing.T) {
	q := makeQueue()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/queue", nil)
	healthcheck.HandleJobQueue(q)(rec, req)

	var body struct {
		Depth int `json:"depth"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Depth != 0 {
		t.Fatalf("expected depth 0, got %d", body.Depth)
	}
}
