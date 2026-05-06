package runlock_test

import (
	"testing"

	"github.com/owner/cronitor-local/internal/runlock"
)

func TestAcquire_FirstCall_Succeeds(t *testing.T) {
	s := runlock.New()
	if !s.Acquire("backup") {
		t.Fatal("expected first acquire to succeed")
	}
}

func TestAcquire_WhileRunning_ReturnsFalse(t *testing.T) {
	s := runlock.New()
	s.Acquire("backup")

	if s.Acquire("backup") {
		t.Fatal("expected second acquire to fail while job is running")
	}
}

func TestAcquire_AfterRelease_Succeeds(t *testing.T) {
	s := runlock.New()
	s.Acquire("backup")
	s.Release("backup")

	if !s.Acquire("backup") {
		t.Fatal("expected acquire to succeed after release")
	}
}

func TestAcquire_DifferentJobs_Independent(t *testing.T) {
	s := runlock.New()
	s.Acquire("job-a")

	if !s.Acquire("job-b") {
		t.Fatal("expected independent jobs to have independent locks")
	}
}

func TestSkips_IncrementOnContention(t *testing.T) {
	s := runlock.New()
	s.Acquire("report")
	s.Acquire("report") // skip 1
	s.Acquire("report") // skip 2

	st, ok := s.Status("report")
	if !ok {
		t.Fatal("expected status to exist")
	}
	if st.Skips != 2 {
		t.Fatalf("expected 2 skips, got %d", st.Skips)
	}
}

func TestStatus_UnknownJob_ReturnsFalse(t *testing.T) {
	s := runlock.New()
	_, ok := s.Status("unknown")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestStatus_RunningFlag(t *testing.T) {
	s := runlock.New()
	s.Acquire("sync")

	st, _ := s.Status("sync")
	if !st.Running {
		t.Fatal("expected running to be true after acquire")
	}

	s.Release("sync")
	st, _ = s.Status("sync")
	if st.Running {
		t.Fatal("expected running to be false after release")
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	s := runlock.New()
	s.Acquire("alpha")
	s.Acquire("beta")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
