package watcher

import (
	"context"
	"time"

	"github.com/yourorg/cronitor-local/internal/logger"
	"github.com/yourorg/cronitor-local/internal/notifier"
	"github.com/yourorg/cronitor-local/internal/runner"
	"github.com/yourorg/cronitor-local/internal/scheduler"
)

// Deps holds the dependencies required by a Watcher.
type Deps struct {
	Scheduler *scheduler.Scheduler
	Runner    *runner.Runner
	Notifier  *notifier.Notifier
	Log       *logger.Logger
	Tick      time.Duration
}

// Watcher polls the scheduler and executes due jobs.
type Watcher struct {
	deps Deps
}

// New returns a Watcher configured with the given dependencies.
func New(deps Deps) *Watcher {
	if deps.Tick == 0 {
		deps.Tick = 30 * time.Second
	}
	return &Watcher{deps: deps}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.deps.Tick)
	defer ticker.Stop()

	w.deps.Log.Infof("watcher started (tick=%s)", w.deps.Tick)

	for {
		select {
		case <-ctx.Done():
			w.deps.Log.Infof("watcher stopped")
			return
		case <-ticker.C:
			w.runDue(ctx)
		}
	}
}

func (w *Watcher) runDue(ctx context.Context) {
	jobs := w.deps.Scheduler.Due()
	for _, job := range jobs {
		job := job
		go w.execute(ctx, job)
	}
}

func (w *Watcher) execute(ctx context.Context, job scheduler.Job) {
	w.deps.Log.Info(job.Name, "starting")
	w.deps.Notifier.Ping(job.Name, notifier.StateRun)

	err := w.deps.Runner.Run(ctx, job.Command)

	if err != nil {
		w.deps.Log.Error(job.Name, err.Error())
		w.deps.Notifier.Ping(job.Name, notifier.StateFail)
		w.deps.Scheduler.RecordError(job.Name)
		return
	}

	w.deps.Log.Info(job.Name, "completed successfully")
	w.deps.Notifier.Ping(job.Name, notifier.StateComplete)
	w.deps.Scheduler.RecordSuccess(job.Name)
}
