package metrics_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronitor-local/internal/metrics"
)

func TestRecordRun_Success(t *testing.T) {
	c := metrics.New()
	c.RecordRun("backup", 2*time.Second, false)

	m, ok := c.Snapshot("backup")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if m.TotalRuns != 1 {
		t.Errorf("TotalRuns: got %d, want 1", m.TotalRuns)
	}
	if m.Failures != 0 {
		t.Errorf("Failures: got %d, want 0", m.Failures)
	}
	if m.LastSuccess.IsZero() {
		t.Error("LastSuccess should be set")
	}
}

func TestRecordRun_Failure(t *testing.T) {
	c := metrics.New()
	c.RecordRun("cleanup", 500*time.Millisecond, true)

	m, ok := c.Snapshot("cleanup")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if m.Failures != 1 {
		t.Errorf("Failures: got %d, want 1", m.Failures)
	}
	if m.LastFailure.IsZero() {
		t.Error("LastFailure should be set")
	}
}

func TestRecordRun_AvgDuration(t *testing.T) {
	c := metrics.New()
	c.RecordRun("report", 2*time.Second, false)
	c.RecordRun("report", 4*time.Second, false)

	m, _ := c.Snapshot("report")
	if m.AvgDuration != 3*time.Second {
		t.Errorf("AvgDuration: got %v, want 3s", m.AvgDuration)
	}
	if m.TotalRuns != 2 {
		t.Errorf("TotalRuns: got %d, want 2", m.TotalRuns)
	}
}

func TestSnapshot_UnknownJob(t *testing.T) {
	c := metrics.New()
	_, ok := c.Snapshot("nonexistent")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	c := metrics.New()
	c.RecordRun("job-a", time.Second, false)
	c.RecordRun("job-b", time.Second, true)

	all := c.All()
	if len(all) != 2 {
		t.Errorf("All: got %d entries, want 2", len(all))
	}
}
