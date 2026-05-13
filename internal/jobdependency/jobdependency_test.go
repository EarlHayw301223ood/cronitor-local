package jobdependency_test

import (
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/jobdependency"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestReady_NoDeps_ReturnsTrue(t *testing.T) {
	s := jobdependency.New(time.Hour)
	ok, _, err := s.Ready("jobA")
	if !ok || err != nil {
		t.Fatalf("expected ready with no deps, got ok=%v err=%v", ok, err)
	}
}

func TestReady_UpstreamNotSeen_ReturnsFalse(t *testing.T) {
	s := jobdependency.New(time.Hour)
	s.Declare("jobB", []string{"jobA"})
	ok, blocking, err := s.Ready("jobB")
	if ok {
		t.Fatal("expected not ready when upstream has never run")
	}
	if blocking != "jobA" {
		t.Errorf("expected blocking=jobA, got %q", blocking)
	}
	if err == nil {
		t.Error("expected non-nil error")
	}
}

func TestReady_UpstreamSucceededWithinWindow_ReturnsTrue(t *testing.T) {
	now := time.Now()
	s := jobdependency.New(time.Hour)
	s.Declare("jobB", []string{"jobA"})
	s.RecordSuccess("jobA")
	ok, _, err := s.Ready("jobB")
	if !ok || err != nil {
		t.Fatalf("expected ready after upstream succeeded, got ok=%v err=%v now=%v", ok, err, now)
	}
}

func TestReady_UpstreamSucceededOutsideWindow_ReturnsFalse(t *testing.T) {
	base := time.Now()
	s := jobdependency.New(time.Minute)

	// Manually set an old success by recording then advancing the clock.
	s.RecordSuccess("jobA")
	// Patch internal clock to two minutes in the future.
	type clockSetter interface{ SetNow(func() time.Time) }
	// We can't patch directly; use a new store with a shifted clock.
	s2 := jobdependency.New(time.Minute)
	_ = base
	s2.Declare("jobB", []string{"jobA"})
	// Simulate: success recorded 2 min ago by using a store whose now() is 2 min ahead.
	// Since we cannot set internal clock externally, we verify via window logic:
	// record success, then create a store with a tiny window to force expiry.
	s3 := jobdependency.New(1 * time.Millisecond)
	s3.Declare("jobC", []string{"jobA"])
	s3.RecordSuccess("jobA")
	time.Sleep(5 * time.Millisecond)
	ok, blocking, _ := s3.Ready("jobC")
	if ok {
		t.Error("expected not ready after window expired")
	}
	if blocking != "jobA" {
		t.Errorf("expected blocking=jobA, got %q", blocking)
	}
}

func TestDeclare_ReplacesOldDeps(t *testing.T) {
	s := jobdependency.New(time.Hour)
	s.Declare("jobB", []string{"jobA", "jobC"})
	s.Declare("jobB", []string{"jobD"})
	all := s.All()
	if len(all["jobB"]) != 1 || all["jobB"][0] != "jobD" {
		t.Errorf("expected only jobD after re-declare, got %v", all["jobB"])
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := jobdependency.New(time.Hour)
	s.Declare("jobB", []string{"jobA"})
	s.Declare("jobC", []string{"jobA", "jobB"})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
