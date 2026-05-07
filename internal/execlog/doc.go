// Package execlog records the stdout/stderr output and exit metadata
// for each job execution, providing a rolling per-job history that can
// be queried via the health-check HTTP API.
//
// Each entry captures the job name, start time, duration, exit code,
// combined output, and whether the run was considered successful.
package execlog
