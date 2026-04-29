package alertmanager_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/cronitor-local/internal/alertmanager"
	"github.com/your-org/cronitor-local/internal/logger"
	"github.com/your-org/cronitor-local/internal/notifier"
	"github.com/your-org/cronitor-local/internal/scheduler"
)

func makeManager(t *testing.T, srv *httptest.Server) (*alertmanager.Manager, *scheduler.Scheduler) {
	t.Helper()
	l := logger.New(nil, logger.Info)
	n := notifier.New(srv.URL+"/%s/%s", &http.Client{}, l)
	s, _ := scheduler.New(l)
	return alertmanager.New(n, s, l), s
}

func TestCheckAndAlert_NoJobs_ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	mgr, _ := makeManager(t, srv)
	alerts := mgr.CheckAndAlert()
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestCheckAndAlert_FailedJob_SendsAlert(t *testing.T) {
	pinged := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pinged = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	mgr, s := makeManager(t, srv)

	_ = s.Register(scheduler.Job{
		Name:     "failing-job",
		Schedule: "* * * * *",
		Command:  "true",
	})
	s.SetLastError("failing-job", errors.New("exit status 1"))

	alerts := mgr.CheckAndAlert()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].JobName != "failing-job" {
		t.Errorf("expected job name 'failing-job', got %q", alerts[0].JobName)
	}
	if !pinged {
		t.Error("expected notifier to be pinged")
	}
}

func TestCheckAndAlert_OverdueJob_SendsMissAlert(t *testing.T) {
	pinged := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pinged = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	mgr, s := makeManager(t, srv)

	_ = s.Register(scheduler.Job{
		Name:     "overdue-job",
		Schedule: "* * * * *",
		Command:  "true",
	})
	s.SetNextRun("overdue-job", time.Now().Add(-5*time.Minute))

	alerts := mgr.CheckAndAlert()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if !pinged {
		t.Error("expected notifier to be pinged for missed job")
	}
}
