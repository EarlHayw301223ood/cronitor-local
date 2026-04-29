package watcher

import (
	"context"
	"log"
	"time"

	"github.com/user/cronitor-local/internal/notifier"
	"github.com/user/cronitor-local/internal/runner"
	"github.com/user/cronitor-local/internal/scheduler"
)

// Watcher orchestrates job execution by listening to the scheduler's
// tick channel and running jobs via the runner, notifying on outcomes.
type Watcher struct {
	sched    *scheduler.Scheduler
	run      *runner.Runner
	notify   *notifier.Notifier
	interval time.Duration
}

// New creates a Watcher that polls the scheduler every interval.
func New(s *scheduler.Scheduler, r *runner.Runner, n *notifier.Notifier, interval time.Duration) *Watcher {
	return &Watcher{
		sched:    s,
		run:      r,
		notify:   n,
		interval: interval,
	}
}

// Start begins the watch loop, blocking until ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("watcher: shutting down")
			return
		case t := <-ticker.C:
			w.tick(ctx, t)
		}
	}
}

func (w *Watcher) tick(ctx context.Context, now time.Time) {
	due := w.sched.Due(now)
	for _, job := range due {
		go w.execute(ctx, job)
	}
}

func (w *Watcher) execute(ctx context.Context, job scheduler.Job) {
	log.Printf("watcher: starting job %q", job.Name)
	w.notify.Ping(job.Name, notifier.StateRun)

	err := w.run.Run(ctx, job.Command)
	if err != nil {
		log.Printf("watcher: job %q failed: %v", job.Name, err)
		w.notify.Ping(job.Name, notifier.StateFail)
		w.sched.RecordError(job.Name, err)
		return
	}

	log.Printf("watcher: job %q succeeded", job.Name)
	w.notify.Ping(job.Name, notifier.StateComplete)
	w.sched.RecordSuccess(job.Name)
}
