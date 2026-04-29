// Package healthcheck provides a simple HTTP server that exposes
// the current status of all monitored cron jobs for external health probes.
package healthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/example/cronitor-local/internal/scheduler"
)

// StatusProvider is satisfied by any type that can return job statuses.
type StatusProvider interface {
	Status() map[string]scheduler.JobStatus
}

// Server is a lightweight HTTP server that serves job health information.
type Server struct {
	scheduler StatusProvider
	port      int
	server    *http.Server
}

// New creates a new healthcheck Server bound to the given port.
func New(s StatusProvider, port int) *Server {
	h := &Server{
		scheduler: s,
		port:      port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/status", h.handleStatus)

	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return h
}

// Start begins listening for HTTP requests. It blocks until the server stops.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop() error {
	return s.server.Close()
}

// handleHealth returns 200 OK if all jobs are healthy, 503 otherwise.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	statuses := s.scheduler.Status()
	for _, st := range statuses {
		if st.LastError != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("unhealthy"))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleStatus returns a JSON summary of all job statuses.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	statuses := s.scheduler.Status()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(statuses)
}
