package jobpriority_test

import (
	"testing"

	"github.com/your-org/cronitor-local/internal/jobpriority"
)

func TestSet_And_Get_ReturnsLevel(t *testing.T) {
	s := jobpriority.New()
	if err := s.Set("backup", jobpriority.High); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if l != jobpriority.High {
		t.Fatalf("want High, got %d", l)
	}
}

func TestGet_UnknownJob_ReturnsNormalAndFalse(t *testing.T) {
	s := jobpriority.New()
	l, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected ok=false for unknown job")
	}
	if l != jobpriority.Normal {
		t.Fatalf("want Normal default, got %d", l)
	}
}

func TestSet_InvalidLevel_ReturnsError(t *testing.T) {
	s := jobpriority.New()
	if err := s.Set("job", jobpriority.Level(99)); err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestSet_OverwritesPreviousLevel(t *testing.T) {
	s := jobpriority.New()
	_ = s.Set("job", jobpriority.Low)
	_ = s.Set("job", jobpriority.High)
	l, _ := s.Get("job")
	if l != jobpriority.High {
		t.Fatalf("want High after overwrite, got %d", l)
	}
}

func TestAll_ReturnsAllTrackedJobs(t *testing.T) {
	s := jobpriority.New()
	_ = s.Set("a", jobpriority.Low)
	_ = s.Set("b", jobpriority.High)
	entries := s.All()
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
}

func TestReset_RemovesEntry(t *testing.T) {
	s := jobpriority.New()
	_ = s.Set("job", jobpriority.High)
	s.Reset("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected entry to be removed after Reset")
	}
}

func TestAll_EmptyStore_ReturnsEmptySlice(t *testing.T) {
	s := jobpriority.New()
	if got := s.All(); len(got) != 0 {
		t.Fatalf("want empty slice, got %d entries", len(got))
	}
}
