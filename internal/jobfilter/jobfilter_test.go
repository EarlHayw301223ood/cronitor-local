package jobfilter_test

import (
	"testing"

	"github.com/your-org/cronitor-local/internal/jobfilter"
)

func TestIsEmpty_NoNames_NoTags(t *testing.T) {
	f := jobfilter.New(nil, nil)
	if !f.IsEmpty() {
		t.Fatal("expected empty filter")
	}
}

func TestIsEmpty_WithNames_NotEmpty(t *testing.T) {
	f := jobfilter.New([]string{"backup"}, nil)
	if f.IsEmpty() {
		t.Fatal("expected non-empty filter")
	}
}

func TestMatchName_EmptyFilter_AcceptsAll(t *testing.T) {
	f := jobfilter.New(nil, nil)
	for _, name := range []string{"a", "b", "backup", ""} {
		if !f.MatchName(name) {
			t.Fatalf("empty filter should accept name %q", name)
		}
	}
}

func TestMatchName_WithNames_AcceptsOnlyListed(t *testing.T) {
	f := jobfilter.New([]string{"backup", "sync"}, nil)
	if !f.MatchName("backup") {
		t.Error("expected backup to match")
	}
	if !f.MatchName("sync") {
		t.Error("expected sync to match")
	}
	if f.MatchName("report") {
		t.Error("expected report not to match")
	}
}

func TestMatchTags_EmptyFilter_AcceptsAll(t *testing.T) {
	f := jobfilter.New(nil, nil)
	if !f.MatchTags([]string{"prod", "nightly"}) {
		t.Fatal("empty filter should accept any tags")
	}
	if !f.MatchTags(nil) {
		t.Fatal("empty filter should accept nil tags")
	}
}

func TestMatchTags_WithTags_MatchesAny(t *testing.T) {
	f := jobfilter.New(nil, []string{"prod"})
	if !f.MatchTags([]string{"staging", "prod"}) {
		t.Error("expected match when one tag overlaps")
	}
	if f.MatchTags([]string{"staging", "dev"}) {
		t.Error("expected no match when no tags overlap")
	}
}

func TestMatch_NameAndTagBothRequired(t *testing.T) {
	f := jobfilter.New([]string{"backup"}, []string{"prod"})
	// correct name, correct tag
	if !f.Match("backup", []string{"prod"}) {
		t.Error("expected match")
	}
	// correct name, wrong tag
	if f.Match("backup", []string{"dev"}) {
		t.Error("expected no match: wrong tag")
	}
	// wrong name, correct tag
	if f.Match("sync", []string{"prod"}) {
		t.Error("expected no match: wrong name")
	}
}

func TestNew_TrimsWhitespace(t *testing.T) {
	f := jobfilter.New([]string{" backup "}, []string{" prod "})
	if !f.MatchName("backup") {
		t.Error("expected trimmed name to match")
	}
	if !f.MatchTags([]string{"prod"}) {
		t.Error("expected trimmed tag to match")
	}
}
