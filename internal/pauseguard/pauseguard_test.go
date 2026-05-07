package pauseguard

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsPaused_NotRegistered_ReturnsFalse(t *testing.T) {
	s := New()
	if s.IsPaused("backup") {
		t.Fatal("expected false for unknown job")
	}
}

func TestPause_Indefinite_ReturnsTrue(t *testing.T) {
	s := New()
	s.Pause("backup", 0)
	if !s.IsPaused("backup") {
		t.Fatal("expected job to be paused indefinitely")
	}
}

func TestPause_WithDuration_PausedBeforeExpiry(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedNow(now)
	s.Pause("backup", 10*time.Minute)
	if !s.IsPaused("backup") {
		t.Fatal("expected job to be paused within duration")
	}
}

func TestPause_WithDuration_NotPausedAfterExpiry(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedNow(now)
	s.Pause("backup", 5*time.Minute)
	// advance clock past expiry
	s.now = fixedNow(now.Add(10 * time.Minute))
	if s.IsPaused("backup") {
		t.Fatal("expected pause to have expired")
	}
}

func TestResume_LiftsPause(t *testing.T) {
	s := New()
	s.Pause("backup", 0)
	s.Resume("backup")
	if s.IsPaused("backup") {
		t.Fatal("expected pause to be lifted after Resume")
	}
}

func TestResume_UnknownJob_IsNoop(t *testing.T) {
	s := New()
	s.Resume("nonexistent") // must not panic
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedNow(now)
	s.Pause("alpha", 0)
	s.Pause("beta", 30*time.Minute)

	statuses := s.All()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	for _, st := range statuses {
		if !st.Paused {
			t.Errorf("expected job %q to be paused", st.Job)
		}
	}
}

func TestAll_ExpiredEntries_NotPaused(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedNow(now)
	s.Pause("gamma", 1*time.Minute)
	s.now = fixedNow(now.Add(2 * time.Minute))

	for _, st := range s.All() {
		if st.Job == "gamma" && st.Paused {
			t.Error("expected expired pause to report Paused=false")
		}
	}
}
