package history_test

import (
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/history"
)

func entry(job string, ok bool) history.Entry {
	return history.Entry{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  100 * time.Millisecond,
		Success:   ok,
		Output:    "output",
	}
}

func TestRecord_AppendsEntry(t *testing.T) {
	h := history.New(10)
	h.Record(entry("job1", true))

	entries := h.Get("job1")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].Success {
		t.Error("expected success=true")
	}
}

func TestRecord_RespectsLimit(t *testing.T) {
	h := history.New(3)
	for i := 0; i < 5; i++ {
		h.Record(entry("job1", i%2 == 0))
	}

	entries := h.Get("job1")
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(entries))
	}
}

func TestGet_UnknownJob_ReturnsEmpty(t *testing.T) {
	h := history.New(10)
	entries := h.Get("unknown")
	if entries == nil {
		t.Fatal("expected non-nil slice for unknown job")
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	h := history.New(10)
	h.Record(entry("job1", true))
	h.Record(entry("job2", false))

	all := h.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(all))
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	h := history.New(10)
	h.Record(entry("job1", true))

	entries := h.Get("job1")
	entries[0].Success = false // mutate the copy

	original := h.Get("job1")
	if !original[0].Success {
		t.Error("mutating returned slice must not affect stored history")
	}
}

func TestNew_DefaultLimit(t *testing.T) {
	h := history.New(0) // should default to 50
	for i := 0; i < 60; i++ {
		h.Record(entry("job1", true))
	}
	if len(h.Get("job1")) != 50 {
		t.Fatalf("expected default limit of 50")
	}
}
