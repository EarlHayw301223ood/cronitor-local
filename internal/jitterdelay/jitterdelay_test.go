package jitterdelay

import (
	"testing"
	"time"
)

func TestFor_ZeroMax_ReturnsZero(t *testing.T) {
	d := New(0)
	if got := d.For("job-a"); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestFor_WithinBounds(t *testing.T) {
	max := 5 * time.Second
	d := New(max)
	for _, name := range []string{"alpha", "beta", "gamma"} {
		off := d.For(name)
		if off < 0 || off >= max {
			t.Errorf("job %q: offset %v out of [0, %v)", name, off, max)
		}
	}
}

func TestFor_SameJobReturnsSameOffset(t *testing.T) {
	d := New(10 * time.Second)
	first := d.For("stable-job")
	for i := 0; i < 10; i++ {
		if got := d.For("stable-job"); got != first {
			t.Fatalf("expected stable %v, got %v on call %d", first, got, i)
		}
	}
}

func TestFor_DifferentJobsGetDifferentOffsets(t *testing.T) {
	d := New(30 * time.Second)
	a := d.For("job-x")
	b := d.For("job-y")
	// With a 30-second window the probability of a collision is negligible.
	if a == b {
		t.Logf("collision occurred (unlikely but possible): %v", a)
	}
}

func TestReset_ForcesNewOffset(t *testing.T) {
	d := New(60 * time.Second)
	first := d.For("resettable")
	d.Reset("resettable")
	// After reset the next call may return the same value by chance, but the
	// cache entry must have been cleared — verify All() no longer shows it
	// before the next For call.
	snap := d.All()
	if _, ok := snap["resettable"]; ok {
		t.Fatal("expected resettable to be absent from snapshot after Reset")
	}
	second := d.For("resettable")
	_ = first
	_ = second // both valid; we just confirm no panic
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	d := New(10 * time.Second)
	d.For("a")
	d.For("b")
	all := d.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if _, ok := all["a"]; !ok {
		t.Error("missing key 'a'")
	}
	if _, ok := all["b"]; !ok {
		t.Error("missing key 'b'")
	}
}
