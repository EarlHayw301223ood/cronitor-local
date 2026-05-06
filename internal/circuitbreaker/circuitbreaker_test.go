package circuitbreaker

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestAllow_ClosedByDefault(t *testing.T) {
	b := New(3, time.Minute)
	if !b.Allow("job1") {
		t.Fatal("expected Allow=true for fresh job")
	}
}

func TestRecordFailure_OpensAfterThreshold(t *testing.T) {
	b := New(3, time.Minute)
	b.RecordFailure("job1")
	b.RecordFailure("job1")
	if b.StateOf("job1") != StateClosed {
		t.Fatal("should still be closed before threshold")
	}
	b.RecordFailure("job1")
	if b.StateOf("job1") != StateOpen {
		t.Fatalf("expected Open, got %v", b.StateOf("job1"))
	}
}

func TestAllow_ReturnsFalseWhenOpen(t *testing.T) {
	b := New(1, time.Minute)
	b.RecordFailure("job1")
	if b.Allow("job1") {
		t.Fatal("expected Allow=false while circuit is open")
	}
}

func TestAllow_HalfOpenAfterCooldown(t *testing.T) {
	now := time.Now()
	b := New(1, time.Minute)
	b.now = fixedNow(now)
	b.RecordFailure("job1")

	// advance past cooldown
	b.now = fixedNow(now.Add(2 * time.Minute))
	if !b.Allow("job1") {
		t.Fatal("expected Allow=true after cooldown (half-open probe)")
	}
	if b.StateOf("job1") != StateHalfOpen {
		t.Fatalf("expected HalfOpen, got %v", b.StateOf("job1"))
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	b := New(1, time.Minute)
	b.RecordFailure("job1")
	b.RecordSuccess("job1")
	if b.StateOf("job1") != StateClosed {
		t.Fatalf("expected Closed after success, got %v", b.StateOf("job1"))
	}
	if !b.Allow("job1") {
		t.Fatal("expected Allow=true after circuit closed")
	}
}

func TestIndependentJobs_DoNotInterfere(t *testing.T) {
	b := New(2, time.Minute)
	b.RecordFailure("jobA")
	b.RecordFailure("jobA")
	if !b.Allow("jobB") {
		t.Fatal("jobB should be unaffected by jobA failures")
	}
	if b.StateOf("jobB") != StateClosed {
		t.Fatalf("jobB should be Closed, got %v", b.StateOf("jobB"))
	}
}

func TestRecordSuccess_ResetsFailureCount(t *testing.T) {
	b := New(3, time.Minute)
	b.RecordFailure("job1")
	b.RecordFailure("job1")
	b.RecordSuccess("job1")
	// one more failure should not trip the breaker (count was reset)
	b.RecordFailure("job1")
	if b.StateOf("job1") != StateClosed {
		t.Fatal("circuit should still be closed after count reset + 1 failure")
	}
}
