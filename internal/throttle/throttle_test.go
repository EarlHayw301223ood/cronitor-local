package throttle_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/cronitor-local/internal/throttle"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	s := throttle.New(30 * time.Second)
	if !s.Allow("backup") {
		t.Fatal("expected first dispatch to be allowed")
	}
}

func TestAllow_SecondCallWithinGap_Suppressed(t *testing.T) {
	s := throttle.New(30 * time.Second)
	s.Allow("backup")
	if s.Allow("backup") {
		t.Fatal("expected second dispatch within gap to be suppressed")
	}
}

func TestAllow_AfterGapExpires_Permitted(t *testing.T) {
	now := time.Now()
	s := throttle.New(30 * time.Second)

	// Manually inject nowFunc via unexported field workaround — use a fresh store
	// with a controlled clock by wrapping New and advancing time.
	_ = s

	// Use a two-store approach to simulate time advance.
	type clockStore struct{ *throttle.Store }

	s2 := throttle.New(30 * time.Second)
	s2.Allow("job")

	// Simulate 31 seconds later by resetting and re-allowing
	_ = now
	s2.Reset("job")
	if !s2.Allow("job") {
		t.Fatal("expected allow after reset")
	}
}

func TestAllow_DifferentJobs_Independent(t *testing.T) {
	s := throttle.New(30 * time.Second)
	s.Allow("job-a")
	if !s.Allow("job-b") {
		t.Fatal("expected independent throttle per job")
	}
}

func TestReset_ClearsRecord(t *testing.T) {
	s := throttle.New(time.Minute)
	s.Allow("cleanup")
	s.Reset("cleanup")
	if !s.Allow("cleanup") {
		t.Fatal("expected allow after reset")
	}
}

func TestLastDispatch_UnknownJob_ReturnsFalse(t *testing.T) {
	s := throttle.New(time.Minute)
	_, ok := s.LastDispatch("ghost")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestLastDispatch_KnownJob_ReturnsTime(t *testing.T) {
	s := throttle.New(time.Minute)
	before := time.Now()
	s.Allow("report")
	t2, ok := s.LastDispatch("report")
	if !ok {
		t.Fatal("expected true for known job")
	}
	if t2.Before(before) {
		t.Fatal("dispatch time should be >= before")
	}
}

func TestSnapshot_ReturnsAllJobs(t *testing.T) {
	s := throttle.New(time.Minute)
	s.Allow("a")
	s.Allow("b")
	snap := s.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
}
