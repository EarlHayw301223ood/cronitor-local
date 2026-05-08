package jobstatus_test

import (
	"testing"
	"time"

	"github.com/example/cronitor-local/internal/jobstatus"
)

func TestSet_And_Get_ReturnsEntry(t *testing.T) {
	s := jobstatus.New()
	s.Set("backup", jobstatus.StatusSuccess)

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != "success" {
		t.Errorf("expected success, got %s", e.Status)
	}
	if e.Job != "backup" {
		t.Errorf("expected job=backup, got %s", e.Job)
	}
}

func TestGet_UnknownJob_ReturnsFalse(t *testing.T) {
	s := jobstatus.New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestSet_OverwritesPreviousStatus(t *testing.T) {
	s := jobstatus.New()
	s.Set("deploy", jobstatus.StatusRunning)
	s.Set("deploy", jobstatus.StatusFailure)

	e, ok := s.Get("deploy")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.Status != "failure" {
		t.Errorf("expected failure, got %s", e.Status)
	}
}

func TestSet_RecordsTimestamp(t *testing.T) {
	before := time.Now()
	s := jobstatus.New()
	s.Set("job", jobstatus.StatusSuccess)
	after := time.Now()

	e, _ := s.Get("job")
	if e.UpdatedAt.Before(before) || e.UpdatedAt.After(after) {
		t.Errorf("timestamp %v not between %v and %v", e.UpdatedAt, before, after)
	}
}

func TestAll_ReturnsAllTrackedJobs(t *testing.T) {
	s := jobstatus.New()
	s.Set("a", jobstatus.StatusSuccess)
	s.Set("b", jobstatus.StatusFailure)
	s.Set("c", jobstatus.StatusRunning)

	all := s.All()
	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}
}

func TestAll_EmptyStore_ReturnsEmptySlice(t *testing.T) {
	s := jobstatus.New()
	all := s.All()
	if all == nil {
		t.Error("expected non-nil slice")
	}
	if len(all) != 0 {
		t.Errorf("expected 0 entries, got %d", len(all))
	}
}

func TestStatus_String(t *testing.T) {
	cases := []struct {
		status jobstatus.Status
		want   string
	}{
		{jobstatus.StatusUnknown, "unknown"},
		{jobstatus.StatusRunning, "running"},
		{jobstatus.StatusSuccess, "success"},
		{jobstatus.StatusFailure, "failure"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}
