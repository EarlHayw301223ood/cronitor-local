package watcher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/config"
	"github.com/your-org/cronitor-local/internal/history"
	"github.com/your-org/cronitor-local/internal/logger"
	"github.com/your-org/cronitor-local/internal/metrics"
	"github.com/your-org/cronitor-local/internal/runner"
	"github.com/your-org/cronitor-local/internal/scheduler"
	"github.com/your-org/cronitor-local/internal/watcher"
)

// integrationDeps wires together real (non-mock) implementations of all
// watcher dependencies so we can exercise the full execution pipeline.
type integrationDeps struct {
	scheduler *scheduler.Scheduler
	runner    *runner.Runner
	history   *history.History
	metrics   *metrics.Metrics
	log       *logger.Logger
}

func newIntegrationDeps(t *testing.T) integrationDeps {
	t.Helper()
	log := logger.New(nil, logger.LevelError) // suppress noise during tests
	return integrationDeps{
		scheduler: scheduler.New(log),
		runner:    runner.New(0),
		history:   history.New(50),
		metrics:   metrics.New(),
		log:       log,
	}
}

// TestIntegration_SuccessfulJobRecordedInHistory verifies that a job which
// exits 0 produces a success entry in both the history and metrics stores.
func TestIntegration_SuccessfulJobRecordedInHistory(t *testing.T) {
	deps := newIntegrationDeps(t)

	job := config.Job{
		Name:     "echo-job",
		Schedule: "* * * * *",
		Command:  "echo hello",
	}
	if err := deps.scheduler.Register(job); err != nil {
		t.Fatalf("register: %v", err)
	}

	w := watcher.New(deps.scheduler, deps.runner, deps.history, deps.metrics, deps.log)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Force the job to be considered due immediately.
	deps.scheduler.MarkDue(job.Name)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	// Give the watcher at least one tick to pick up the job.
	time.Sleep(300 * time.Millisecond)
	cancel()
	wg.Wait()

	entries := deps.history.Get(job.Name)
	if len(entries) == 0 {
		t.Fatal("expected at least one history entry, got none")
	}
	if entries[0].Error != "" {
		t.Errorf("expected no error, got %q", entries[0].Error)
	}

	snap := deps.metrics.Snapshot(job.Name)
	if snap.Runs == 0 {
		t.Error("expected metrics to record at least one run")
	}
	if snap.Failures != 0 {
		t.Errorf("expected zero failures, got %d", snap.Failures)
	}
}

// TestIntegration_FailingJobRecordedAsFailure verifies that a job which
// exits non-zero increments the failure counter in metrics.
func TestIntegration_FailingJobRecordedAsFailure(t *testing.T) {
	deps := newIntegrationDeps(t)

	job := config.Job{
		Name:     "fail-job",
		Schedule: "* * * * *",
		Command:  "false", // always exits 1
	}
	if err := deps.scheduler.Register(job); err != nil {
		t.Fatalf("register: %v", err)
	}

	w := watcher.New(deps.scheduler, deps.runner, deps.history, deps.metrics, deps.log)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	deps.scheduler.MarkDue(job.Name)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	time.Sleep(300 * time.Millisecond)
	cancel()
	wg.Wait()

	snap := deps.metrics.Snapshot(job.Name)
	if snap.Failures == 0 {
		t.Error("expected at least one failure to be recorded")
	}

	entries := deps.history.Get(job.Name)
	if len(entries) == 0 {
		t.Fatal("expected at least one history entry")
	}
	if entries[0].Error == "" {
		t.Error("expected history entry to contain an error message")
	}
}
