package watcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/cronitor-local/internal/notifier"
	"github.com/user/cronitor-local/internal/runner"
	"github.com/user/cronitor-local/internal/scheduler"
	"github.com/user/cronitor-local/internal/watcher"
)

func makeTestDeps(t *testing.T, pingCounter *atomic.Int32) (*scheduler.Scheduler, *runner.Runner, *notifier.Notifier) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingCounter.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	sched := scheduler.New()
	r := runner.New(5 * time.Second)
	n := notifier.New(server.URL, "test-key")
	return sched, r, n
}

func TestWatcher_ExecutesDueJob(t *testing.T) {
	var pings atomic.Int32
	sched, r, n := makeTestDeps(t, &pings)

	_ = sched.Register(scheduler.Job{
		Name:     "echo-job",
		Command:  "echo hello",
		Schedule: "* * * * *",
	})

	w := watcher.New(sched, r, n, 50*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w.Start(ctx)

	if pings.Load() == 0 {
		t.Error("expected at least one ping, got none")
	}
}

func TestWatcher_StopsOnContextCancel(t *testing.T) {
	var pings atomic.Int32
	sched, r, n := makeTestDeps(t, &pings)

	w := watcher.New(sched, r, n, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher did not stop after context cancellation")
	}
}
