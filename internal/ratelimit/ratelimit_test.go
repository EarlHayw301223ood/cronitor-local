package ratelimit_test

import (
	"testing"
	"time"

	"github.com/example/cronitor-local/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(time.Minute)
	if !l.Allow("backup") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinCooldown_Suppressed(t *testing.T) {
	l := ratelimit.New(time.Minute)
	l.Allow("backup")
	if l.Allow("backup") {
		t.Fatal("expected second call within cooldown to be suppressed")
	}
}

func TestAllow_AfterCooldownExpires_Permitted(t *testing.T) {
	l := ratelimit.New(10 * time.Millisecond)
	l.Allow("backup")
	time.Sleep(20 * time.Millisecond)
	if !l.Allow("backup") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_DifferentJobs_IndependentLimits(t *testing.T) {
	l := ratelimit.New(time.Minute)
	l.Allow("job-a")
	if !l.Allow("job-b") {
		t.Fatal("expected independent jobs to have independent limits")
	}
}

func TestReset_ClearsRecord(t *testing.T) {
	l := ratelimit.New(time.Minute)
	l.Allow("backup")
	l.Reset("backup")
	if !l.Allow("backup") {
		t.Fatal("expected call after reset to be allowed")
	}
}

func TestSnapshot_ReturnsCurrentState(t *testing.T) {
	l := ratelimit.New(time.Minute)
	l.Allow("job-a")
	l.Allow("job-b")
	snap := l.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries in snapshot, got %d", len(snap))
	}
	if _, ok := snap["job-a"]; !ok {
		t.Error("expected job-a in snapshot")
	}
	if _, ok := snap["job-b"]; !ok {
		t.Error("expected job-b in snapshot")
	}
}

func TestSnapshot_IsCopy(t *testing.T) {
	l := ratelimit.New(time.Minute)
	l.Allow("backup")
	snap := l.Snapshot()
	delete(snap, "backup")
	if len(l.Snapshot()) != 1 {
		t.Fatal("modifying snapshot must not affect internal state")
	}
}
