package jobtimeout_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronitor-local/internal/jobtimeout"
)

func TestSet_And_Get_ReturnsTimeout(t *testing.T) {
	s := jobTimeout.New()
	if err := s.Set("backup", 30*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if d != 30*time.Second {
		t.Fatalf("want 30s, got %v", d)
	}
}

func TestGet_UnknownJob_ReturnsFalse(t *testing.T) {
	s := jobTimeout.New()
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestSet_InvalidTimeout_ReturnsError(t *testing.T) {
	s := jobTimeout.New()
	if err := s.Set("job", 0); err == nil {
		t.Fatal("expected error for zero duration")
	}
	if err := s.Set("job", -1*time.Second); err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestSet_OverwritesPreviousTimeout(t *testing.T) {
	s := jobTimeout.New()
	_ = s.Set("job", 10*time.Second)
	_ = s.Set("job", 60*time.Second)
	d, _ := s.Get("job")
	if d != 60*time.Second {
		t.Fatalf("want 60s, got %v", d)
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s := jobTimeout.New()
	_ = s.Set("job", 5*time.Second)
	s.Delete("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	s := jobTimeout.New()
	_ = s.Set("alpha", 10*time.Second)
	_ = s.Set("beta", 20*time.Second)
	entries := s.All()
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
}

func TestAll_EmptyStore_ReturnsEmptySlice(t *testing.T) {
	s := jobTimeout.New()
	if entries := s.All(); len(entries) != 0 {
		t.Fatalf("want empty slice, got %d entries", len(entries))
	}
}
