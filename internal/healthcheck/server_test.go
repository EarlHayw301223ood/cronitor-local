package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronitor-local/internal/history"
	"github.com/user/cronitor-local/internal/metrics"
	"github.com/user/cronitor-local/internal/scheduler"
)

func makeServer(t *testing.T) *Server {
	t.Helper()
	sched := scheduler.New()
	hist := history.New(50)
	met := metrics.New()
	return New(sched, hist, met, nil)
}

func TestNew_ReturnsServer(t *testing.T) {
	s := makeServer(t)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestRegister_MountsAllRoutes(t *testing.T) {
	s := makeServer(t)
	mux := http.NewServeMux()
	s.Register(mux)

	routes := []string{"/health", "/status", "/metrics", "/history", "/jobs"}
	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// Any non-404 response means the route is registered.
		if w.Code == http.StatusNotFound {
			t.Errorf("route %s not registered (got 404)", route)
		}
	}
}

func TestRegister_HealthReturns200(t *testing.T) {
	s := makeServer(t)
	mux := http.NewServeMux()
	s.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from /health, got %d", w.Code)
	}
}

func TestRegister_MetricsReturns200(t *testing.T) {
	s := makeServer(t)
	mux := http.NewServeMux()
	s.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from /metrics, got %d", w.Code)
	}
}
