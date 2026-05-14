package jobenv_test

import (
	"testing"

	"github.com/yourorg/cronitor-local/internal/jobenv"
)

func TestSet_And_Get_ReturnsEnv(t *testing.T) {
	s := jobenv.New()
	env := map[string]string{"FOO": "bar", "BAZ": "qux"}
	if err := s.Set("myjob", env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("myjob")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got["FOO"] != "bar" || got["BAZ"] != "qux" {
		t.Errorf("unexpected env: %v", got)
	}
}

func TestGet_UnknownJob_ReturnsFalse(t *testing.T) {
	s := jobenv.New()
	_, ok := s.Get("ghost")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestSet_EmptyJobName_ReturnsError(t *testing.T) {
	s := jobenv.New()
	if err := s.Set("", map[string]string{"K": "V"}); err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestSet_OverwritesPreviousEnv(t *testing.T) {
	s := jobenv.New()
	_ = s.Set("job", map[string]string{"A": "1"})
	_ = s.Set("job", map[string]string{"B": "2"})
	got, _ := s.Get("job")
	if _, exists := got["A"]; exists {
		t.Error("old key should have been replaced")
	}
	if got["B"] != "2" {
		t.Errorf("expected B=2, got %v", got)
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s := jobenv.New()
	_ = s.Set("job", map[string]string{"X": "y"})
	s.Delete("job")
	_, ok := s.Get("job")
	if ok {
		t.Error("expected entry to be deleted")
	}
}

func TestAll_ReturnsAllTrackedJobs(t *testing.T) {
	s := jobenv.New()
	_ = s.Set("alpha", map[string]string{"K": "1"})
	_ = s.Set("beta", map[string]string{"K": "2"})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(all))
	}
}

func TestGet_ReturnsCopy_MutationDoesNotAffectStore(t *testing.T) {
	s := jobenv.New()
	_ = s.Set("job", map[string]string{"K": "original"})
	got, _ := s.Get("job")
	got["K"] = "mutated"
	again, _ := s.Get("job")
	if again["K"] != "original" {
		t.Error("store was mutated through returned map")
	}
}
