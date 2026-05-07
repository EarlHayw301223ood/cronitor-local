package overlapguard_test

import (
	"testing"

	"github.com/your-org/cronitor-local/internal/overlapguard"
)

func TestEnter_FirstCall_Succeeds(t *testing.T) {
	g := overlapguard.New()
	if !g.Enter("backup") {
		t.Fatal("expected Enter to return true for a fresh job")
	}
}

func TestEnter_WhileRunning_ReturnsFalse(t *testing.T) {
	g := overlapguard.New()
	g.Enter("backup")
	if g.Enter("backup") {
		t.Fatal("expected Enter to return false while job is still running")
	}
}

func TestEnter_WhileRunning_IncrementsSkips(t *testing.T) {
	g := overlapguard.New()
	g.Enter("backup")
	g.Enter("backup")
	g.Enter("backup")
	if got := g.Skips("backup"); got != 2 {
		t.Fatalf("expected 2 skips, got %d", got)
	}
}

func TestExit_AllowsReEntry(t *testing.T) {
	g := overlapguard.New()
	g.Enter("backup")
	g.Exit("backup")
	if !g.Enter("backup") {
		t.Fatal("expected Enter to succeed after Exit")
	}
}

func TestExit_UnknownJob_IsNoop(t *testing.T) {
	g := overlapguard.New()
	// should not panic
	g.Exit("nonexistent")
}

func TestDifferentJobs_Independent(t *testing.T) {
	g := overlapguard.New()
	g.Enter("jobA")
	if !g.Enter("jobB") {
		t.Fatal("jobB should be independent of jobA")
	}
}

func TestIsRunning_TrueWhileActive(t *testing.T) {
	g := overlapguard.New()
	g.Enter("sync")
	if !g.IsRunning("sync") {
		t.Fatal("expected IsRunning to be true after Enter")
	}
	g.Exit("sync")
	if g.IsRunning("sync") {
		t.Fatal("expected IsRunning to be false after Exit")
	}
}

func TestSnapshot_ReturnsAllSkips(t *testing.T) {
	g := overlapguard.New()
	g.Enter("a")
	g.Enter("a") // skip
	g.Enter("b")
	g.Enter("b") // skip
	g.Enter("b") // skip

	snap := g.Snapshot()
	if snap["a"] != 1 {
		t.Fatalf("expected a=1, got %d", snap["a"])
	}
	if snap["b"] != 2 {
		t.Fatalf("expected b=2, got %d", snap["b"])
	}
}

func TestSkips_UnknownJob_ReturnsZero(t *testing.T) {
	g := overlapguard.New()
	if got := g.Skips("ghost"); got != 0 {
		t.Fatalf("expected 0 skips for unknown job, got %d", got)
	}
}
