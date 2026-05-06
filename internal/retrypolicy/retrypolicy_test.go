package retrypolicy_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronitor-local/internal/retrypolicy"
)

func TestShouldRetry_WithinBudget(t *testing.T) {
	p := retrypolicy.New(3, time.Second)
	for i := 0; i < 3; i++ {
		if !p.ShouldRetry("job1") {
			t.Fatalf("expected retry to be allowed on attempt %d", i+1)
		}
	}
}

func TestShouldRetry_ExhaustedBudget(t *testing.T) {
	p := retrypolicy.New(2, time.Second)
	p.ShouldRetry("job1")
	p.ShouldRetry("job1")
	if p.ShouldRetry("job1") {
		t.Fatal("expected retry to be denied after budget exhausted")
	}
}

func TestShouldRetry_IndependentJobs(t *testing.T) {
	p := retrypolicy.New(1, time.Second)
	p.ShouldRetry("jobA")
	if !p.ShouldRetry("jobB") {
		t.Fatal("jobB should still have retry budget independent of jobA")
	}
}

func TestReset_ClearsAttempts(t *testing.T) {
	p := retrypolicy.New(1, time.Second)
	p.ShouldRetry("job1") // exhaust budget
	if p.ShouldRetry("job1") {
		t.Fatal("budget should be exhausted before reset")
	}
	p.Reset("job1")
	if !p.ShouldRetry("job1") {
		t.Fatal("budget should be restored after reset")
	}
}

func TestAttempts_TracksCount(t *testing.T) {
	p := retrypolicy.New(5, time.Second)
	p.ShouldRetry("job1")
	p.ShouldRetry("job1")
	if got := p.Attempts("job1"); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}

func TestBackoff_ExponentialGrowth(t *testing.T) {
	base := 100 * time.Millisecond
	p := retrypolicy.New(5, base)

	// first attempt recorded
	p.ShouldRetry("job1")
	if got := p.Backoff("job1"); got != base {
		t.Fatalf("expected %v after 1 attempt, got %v", base, got)
	}

	// second attempt
	p.ShouldRetry("job1")
	want := 2 * base
	if got := p.Backoff("job1"); got != want {
		t.Fatalf("expected %v after 2 attempts, got %v", want, got)
	}
}

func TestBackoff_NoAttempts_ReturnsBase(t *testing.T) {
	base := 500 * time.Millisecond
	p := retrypolicy.New(3, base)
	if got := p.Backoff("unknown"); got != base {
		t.Fatalf("expected base backoff %v, got %v", base, got)
	}
}
