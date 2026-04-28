package scheduler

import (
	"errors"
	"testing"
	"time"

	"cronitor-local/internal/config"
)

func makeJobs() []config.Job {
	return []config.Job{
		{Name: "test-job", Schedule: "@every 1s", Command: "echo hello"},
	}
}

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if s.cron == nil {
		t.Fatal("expected non-nil cron instance")
	}
}

func TestRegister_ValidJob(t *testing.T) {
	s := New()
	called := make(chan struct{}, 1)

	runner := func(j config.Job) error {
		called <- struct{}{}
		return nil
	}

	err := s.Register(makeJobs(), runner)
	if err != nil {
		t.Fatalf("unexpected error registering job: %v", err)
	}

	s.Start()
	defer s.Stop()

	select {
	case <-called:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("job was not executed within timeout")
	}
}

func TestRegister_InvalidSchedule(t *testing.T) {
	s := New()
	jobs := []config.Job{
		{Name: "bad-job", Schedule: "not-a-cron", Command: "echo hi"},
	}
	err := s.Register(jobs, func(j config.Job) error { return nil })
	if err == nil {
		t.Fatal("expected error for invalid schedule, got nil")
	}
}

func TestStatus_TracksRunAndError(t *testing.T) {
	s := New()
	jobErr := errors.New("simulated failure")
	done := make(chan struct{}, 1)

	runner := func(j config.Job) error {
		done <- struct{}{}
		return jobErr
	}

	err := s.Register(makeJobs(), runner)
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	s.Start()
	defer s.Stop()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("job did not run within timeout")
	}

	time.Sleep(50 * time.Millisecond) // allow status update
	statuses := s.Status()
	st, ok := statuses["test-job"]
	if !ok {
		t.Fatal("expected status entry for test-job")
	}
	if st.RunCount < 1 {
		t.Errorf("expected RunCount >= 1, got %d", st.RunCount)
	}
	if st.LastError == nil {
		t.Error("expected LastError to be set")
	}
}
