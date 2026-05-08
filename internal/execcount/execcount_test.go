package execcount_test

import (
	"testing"

	"github.com/example/cronitor-local/internal/execcount"
)

func TestRecord_Success_IncrementsTotal(t *testing.T) {
	s := execcount.New()
	s.Record("backup", true)

	c := s.Get("backup")
	if c.Total != 1 {
		t.Fatalf("expected Total=1, got %d", c.Total)
	}
	if c.Success != 1 {
		t.Fatalf("expected Success=1, got %d", c.Success)
	}
	if c.Failures != 0 {
		t.Fatalf("expected Failures=0, got %d", c.Failures)
	}
}

func TestRecord_Failure_IncrementsFailures(t *testing.T) {
	s := execcount.New()
	s.Record("backup", false)

	c := s.Get("backup")
	if c.Total != 1 {
		t.Fatalf("expected Total=1, got %d", c.Total)
	}
	if c.Failures != 1 {
		t.Fatalf("expected Failures=1, got %d", c.Failures)
	}
	if c.Success != 0 {
		t.Fatalf("expected Success=0, got %d", c.Success)
	}
}

func TestRecord_MultipleRuns_AccumulatesCorrectly(t *testing.T) {
	s := execcount.New()
	s.Record("sync", true)
	s.Record("sync", true)
	s.Record("sync", false)

	c := s.Get("sync")
	if c.Total != 3 {
		t.Fatalf("expected Total=3, got %d", c.Total)
	}
	if c.Success != 2 {
		t.Fatalf("expected Success=2, got %d", c.Success)
	}
	if c.Failures != 1 {
		t.Fatalf("expected Failures=1, got %d", c.Failures)
	}
}

func TestGet_UnknownJob_ReturnsZeroCounter(t *testing.T) {
	s := execcount.New()
	c := s.Get("nonexistent")
	if c.Total != 0 || c.Success != 0 || c.Failures != 0 {
		t.Fatalf("expected zero counter, got %+v", c)
	}
}

func TestAll_ReturnsAllTrackedJobs(t *testing.T) {
	s := execcount.New()
	s.Record("jobA", true)
	s.Record("jobB", false)

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(all))
	}
	if all["jobA"].Success != 1 {
		t.Errorf("jobA success mismatch")
	}
	if all["jobB"].Failures != 1 {
		t.Errorf("jobB failure mismatch")
	}
}

func TestReset_ClearsCounters(t *testing.T) {
	s := execcount.New()
	s.Record("cleanup", true)
	s.Record("cleanup", false)
	s.Reset("cleanup")

	c := s.Get("cleanup")
	if c.Total != 0 {
		t.Fatalf("expected Total=0 after reset, got %d", c.Total)
	}
}

func TestAll_IsolatedSnapshot(t *testing.T) {
	s := execcount.New()
	s.Record("jobA", true)

	snap := s.All()
	s.Record("jobA", false) // mutate after snapshot

	if snap["jobA"].Total != 1 {
		t.Fatalf("snapshot should not reflect later mutations, got Total=%d", snap["jobA"].Total)
	}
}
