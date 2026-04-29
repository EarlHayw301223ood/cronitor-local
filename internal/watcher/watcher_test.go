package watcher_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronitor-local/internal/logger"
	"github.com/yourorg/cronitor-local/internal/notifier"
	"github.com/yourorg/cronitor-local/internal/runner"
	"github.com/yourorg/cronitor-local/internal/scheduler"
	"github.com/yourorg/cronitor-local/internal/watcher"
)

func makeTestDeps(t *testing.T, serverURL string) watcher.Deps {
	t.Helper()
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelDebug)
	sched := scheduler.New()
	run := runner.New(runner.Options{Timeout: 5 * time.Second})
	not := notifier.New(serverURL)
	return watcher.Deps{
		Scheduler: sched,
		Runner:    run,
		Notifier:  not,
		Log:       log,
		Tick:      50 * time.Millisecond,
	}
}

func TestWatcher_ExecutesDueJob(t *testing.T) {
	pinged := make(chan string, 10)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pinged <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	deps := makeTestDeps(t, srv.URL)
	err := deps.Scheduler.Register(scheduler.Job{
		Name:    "echo-job",
		Command: "echo hello",
		Schedule: "* * * * *",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w := watcher.New(deps)
	go w.Run(ctx)

	select {
	case path := <-pinged:
		if path == "" {
			t.Error("expected non-empty ping path")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out waiting for ping")
	}
}

func TestWatcher_StopsOnContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	deps := makeTestDeps(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	w := watcher.New(deps)
	go func() {
		w.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Error("watcher did not stop after context cancel")
	}
}
