package runner

import (
	"context"
	"os/exec"
	"time"
)

// Result holds the outcome of a single job execution.
type Result struct {
	JobName  string
	Output   string
	Err      error
	Duration time.Duration
	Started  time.Time
}

// Runner executes shell commands for scheduled jobs.
type Runner struct {
	timeout time.Duration
}

// New creates a Runner with the given execution timeout.
func New(timeout time.Duration) *Runner {
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	return &Runner{timeout: timeout}
}

// Run executes the given shell command for jobName and returns a Result.
// The command is run via "sh -c" to support pipes and shell features.
// Execution is cancelled if it exceeds the configured timeout.
func (r *Runner) Run(jobName, command string) Result {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		err = context.DeadlineExceeded
	}

	return Result{
		JobName:  jobName,
		Output:   string(out),
		Err:      err,
		Duration: time.Since(start),
		Started:  start,
	}
}
