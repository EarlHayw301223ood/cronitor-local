package jobqueue_test

import (
	"testing"

	"github.com/yourorg/cronitor-local/internal/jobqueue"
)

func TestEnqueue_AddsEntry(t *testing.T) {
	q := jobqueue.New(4)
	q.Enqueue("backup")
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestDequeue_ReturnsOldestEntry(t *testing.T) {
	q := jobqueue.New(4)
	q.Enqueue("first")
	q.Enqueue("second")

	e, ok := q.Dequeue()
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if e.JobName != "first" {
		t.Fatalf("expected 'first', got %q", e.JobName)
	}
}

func TestDequeue_EmptyQueue_ReturnsFalse(t *testing.T) {
	q := jobqueue.New(4)
	_, ok := q.Dequeue()
	if ok {
		t.Fatal("expected false for empty queue")
	}
}

func TestEnqueue_EvictsOldestWhenFull(t *testing.T) {
	q := jobqueue.New(2)
	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c") // evicts "a"

	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
	e, _ := q.Dequeue()
	if e.JobName != "b" {
		t.Fatalf("expected 'b' at head, got %q", e.JobName)
	}
}

func TestEvictions_TracksEvictedJobs(t *testing.T) {
	q := jobqueue.New(1)
	q.Enqueue("victim")
	q.Enqueue("new") // evicts "victim"

	if q.Evictions("victim") != 1 {
		t.Fatalf("expected 1 eviction for 'victim', got %d", q.Evictions("victim"))
	}
}

func TestAllEvictions_ReturnsAllJobs(t *testing.T) {
	q := jobqueue.New(1)
	q.Enqueue("a")
	q.Enqueue("b") // evicts a
	q.Enqueue("c") // evicts b

	all := q.AllEvictions()
	if all["a"] != 1 {
		t.Fatalf("expected 1 eviction for a, got %d", all["a"])
	}
	if all["b"] != 1 {
		t.Fatalf("expected 1 eviction for b, got %d", all["b"])
	}
}

func TestNew_DefaultLimit_UsedWhenZero(t *testing.T) {
	q := jobqueue.New(0)
	for i := 0; i < 64; i++ {
		q.Enqueue("job")
	}
	if q.Len() != 64 {
		t.Fatalf("expected len 64, got %d", q.Len())
	}
}
