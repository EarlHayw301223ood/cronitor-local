package execlog_test

import (
	"testing"
	"time"

	"github.com/example/cronitor-local/internal/execlog"
)

func entry(job string, code int, output string) execlog.Entry {
	return execlog.Entry{
		JobName:   job,
		StartedAt: time.Now(),
		Output:    output,
		ExitCode:  code,
	}
}

func TestRecord_AppendsEntry(t *testing.T) {
	s := execlog.New(5)
	s.Record(entry("backup", 0, "ok"))

	got := s.Get("backup")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", got[0].ExitCode)
	}
}

func TestRecord_RespectsLimit(t *testing.T) {
	s := execlog.New(3)
	for i := 0; i < 5; i++ {
		s.Record(entry("backup", i, ""))
	}

	got := s.Get("backup")
	if len(got) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(got))
	}
	// Oldest should have been evicted; first remaining exit code == 2.
	if got[0].ExitCode != 2 {
		t.Errorf("expected oldest remaining exit code 2, got %d", got[0].ExitCode)
	}
}

func TestGet_UnknownJob_ReturnsEmpty(t *testing.T) {
	s := execlog.New(5)
	got := s.Get("nonexistent")
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	s := execlog.New(5)
	s.Record(entry("jobA", 0, "a"))
	s.Record(entry("jobB", 1, "b"))

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(all))
	}
	if len(all["jobA"]) != 1 || len(all["jobB"]) != 1 {
		t.Error("expected one entry per job")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := execlog.New(5)
	s.Record(entry("jobA", 0, "original"))

	all := s.All()
	all["jobA"][0].Output = "mutated"

	again := s.Get("jobA")
	if again[0].Output != "original" {
		t.Error("All() should return an independent copy")
	}
}
