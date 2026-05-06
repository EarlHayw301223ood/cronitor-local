package deadletter_test

import (
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/deadletter"
)

func entry(job, cmd, errMsg string, attempts int) deadletter.Entry {
	return deadletter.Entry{
		JobName:  job,
		Command:  cmd,
		Error:    errMsg,
		Attempts: attempts,
	}
}

func TestAdd_AppendsEntry(t *testing.T) {
	s := deadletter.New(10)
	s.Add(entry("backup", "backup.sh", "exit 1", 3))
	if s.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", s.Len())
	}
}

func TestAdd_SetsCreatedAt(t *testing.T) {
	s := deadletter.New(10)
	before := time.Now()
	s.Add(entry("backup", "backup.sh", "timeout", 2))
	after := time.Now()
	all := s.All()
	if all[0].CreatedAt.Before(before) || all[0].CreatedAt.After(after) {
		t.Error("CreatedAt not set to approximately now")
	}
}

func TestAdd_EvictsOldestWhenLimitReached(t *testing.T) {
	s := deadletter.New(3)
	s.Add(entry("job-a", "a.sh", "err", 1))
	s.Add(entry("job-b", "b.sh", "err", 1))
	s.Add(entry("job-c", "c.sh", "err", 1))
	s.Add(entry("job-d", "d.sh", "err", 1))
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
	if all[0].JobName != "job-b" {
		t.Errorf("expected oldest evicted, first entry should be job-b, got %s", all[0].JobName)
	}
}

func TestForJob_FiltersCorrectly(t *testing.T) {
	s := deadletter.New(20)
	s.Add(entry("alpha", "a.sh", "err", 2))
	s.Add(entry("beta", "b.sh", "err", 2))
	s.Add(entry("alpha", "a.sh", "err", 3))

	result := s.ForJob("alpha")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries for alpha, got %d", len(result))
	}
	for _, e := range result {
		if e.JobName != "alpha" {
			t.Errorf("unexpected job name %s", e.JobName)
		}
	}
}

func TestForJob_UnknownJob_ReturnsNil(t *testing.T) {
	s := deadletter.New(10)
	s.Add(entry("alpha", "a.sh", "err", 1))
	if got := s.ForJob("ghost"); got != nil {
		t.Errorf("expected nil for unknown job, got %v", got)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := deadletter.New(10)
	s.Add(entry("job", "j.sh", "err", 1))
	a := s.All()
	a[0].JobName = "mutated"
	if s.All()[0].JobName == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}
