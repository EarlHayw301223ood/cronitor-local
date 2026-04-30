// Package healthcheck provides an HTTP server exposing health, status,
// metrics, history, and job endpoints for cronitor-local.
package healthcheck

import (
	"fmt"
	"net/http"

	"github.com/user/cronitor-local/internal/history"
	"github.com/user/cronitor-local/internal/logger"
	"github.com/user/cronitor-local/internal/metrics"
	"github.com/user/cronitor-local/internal/scheduler"
)

// Server holds dependencies for all HTTP handlers.
type Server struct {
	scheduler *scheduler.Scheduler
	history   *history.Store
	metrics   *metrics.Store
	log       *logger.Logger
}

// New creates a Server wiring up all handler dependencies.
func New(
	sched *scheduler.Scheduler,
	hist *history.Store,
	met *metrics.Store,
	log *logger.Logger,
) *Server {
	return &Server{
		scheduler: sched,
		history:   hist,
		metrics:   met,
		log:       log,
	}
}

// Register mounts all routes onto the provided mux.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/status", s.HandleStatus)
	mux.HandleFunc("/metrics", s.HandleMetrics)
	mux.HandleFunc("/history", s.HandleHistory)
	mux.HandleFunc("/jobs", s.HandleJobs)
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	s.Register(mux)
	if s.log != nil {
		s.log.Info("", fmt.Sprintf("healthcheck server listening on %s", addr))
	}
	return http.ListenAndServe(addr, mux)
}
