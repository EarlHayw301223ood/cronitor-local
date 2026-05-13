package jobtagindex_test

import (
	"sort"
	"testing"

	"github.com/example/cronitor-local/internal/jobtagindex"
)

func sorted(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

func TestRegister_AssociatesTagsWithJob(t *testing.T) {
	idx := jobtagindex.New()
	idx.Register("backup", []string{"daily", "critical"})

	got := sorted(idx.JobsForTag("daily"))
	if len(got) != 1 || got[0] != "backup" {
		t.Fatalf("expected [backup], got %v", got)
	}
}

func TestRegister_MultipleJobsSameTag(t *testing.T) {
	idx := jobtagindex.New()
	idx.Register("backup", []string{"daily"})
	idx.Register("report", []string{"daily"})

	got := sorted(idx.JobsForTag("daily"))
	if len(got) != 2 || got[0] != "backup" || got[1] != "report" {
		t.Fatalf("expected [backup report], got %v", got)
	}
}

func TestRegister_ReplacesOldTags(t *testing.T) {
	idx := jobtagindex.New()
	idx.Register("backup", []string{"daily", "critical"})
	idx.Register("backup", []string{"weekly"})

	if jobs := idx.JobsForTag("daily"); len(jobs) != 0 {
		t.Fatalf("expected old tag 'daily' removed, got %v", jobs)
	}
	got := sorted(idx.JobsForTag("weekly"))
	if len(got) != 1 || got[0] != "backup" {
		t.Fatalf("expected [backup] under 'weekly', got %v", got)
	}
}

func TestJobsForTag_UnknownTag_ReturnsNil(t *testing.T) {
	idx := jobtagindex.New()
	if jobs := idx.JobsForTag("nope"); jobs != nil {
		t.Fatalf("expected nil, got %v", jobs)
	}
}

func TestTags_ReturnsAllTrackedTags(t *testing.T) {
	idx := jobtagindex.New()
	idx.Register("job1", []string{"alpha", "beta"})
	idx.Register("job2", []string{"gamma"})

	tags := sorted(idx.Tags())
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %v", tags)
	}
}

func TestDeregister_RemovesJob(t *testing.T) {
	idx := jobtagindex.New()
	idx.Register("backup", []string{"daily"})
	idx.Deregister("backup")

	if jobs := idx.JobsForTag("daily"); len(jobs) != 0 {
		t.Fatalf("expected empty after deregister, got %v", jobs)
	}
	if tags := idx.Tags(); len(tags) != 0 {
		t.Fatalf("expected no tags after deregister, got %v", tags)
	}
}

func TestDeregister_UnknownJob_IsNoop(t *testing.T) {
	idx := jobtagindex.New()
	idx.Deregister("ghost") // must not panic
}
