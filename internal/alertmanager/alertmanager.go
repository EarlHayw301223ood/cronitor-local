// Package alertmanager coordinates sending alerts when jobs miss
// their expected execution window or exit with a non-zero status.
package alertmanager

import (
	"fmt"
	"time"

	"github.com/your-org/cronitor-local/internal/logger"
	"github.com/your-org/cronitor-local/internal/notifier"
	"github.com/your-org/cronitor-local/internal/scheduler"
)

// Alert represents a single alert event.
type Alert struct {
	JobName   string
	Reason    string
	OccurredAt time.Time
}

// Manager checks job statuses and fires alerts via the notifier.
type Manager struct {
	notifier  *notifier.Notifier
	scheduler *scheduler.Scheduler
	log       *logger.Logger
}

// New creates a new Manager.
func New(n *notifier.Notifier, s *scheduler.Scheduler, l *logger.Logger) *Manager {
	return &Manager{
		notifier:  n,
		scheduler: s,
		log:       l,
	}
}

// CheckAndAlert inspects all job statuses and sends an alert for any
// job that has failed or has not run within its expected window.
func (m *Manager) CheckAndAlert() []Alert {
	statuses := m.scheduler.Status()
	var fired []Alert

	for name, st := range statuses {
		if st.LastError != nil {
			reason := fmt.Sprintf("job %q failed: %v", name, st.LastError)
			m.log.Error(nil, reason)
			m.notifier.Ping(name, "fail")
			fired = append(fired, Alert{
				JobName:    name,
				Reason:     reason,
				OccurredAt: time.Now(),
			})
			continue
		}

		if st.NextRun.Before(time.Now()) && st.LastRun.IsZero() {
			reason := fmt.Sprintf("job %q has never run and is overdue", name)
			m.log.Error(nil, reason)
			m.notifier.Ping(name, "miss")
			fired = append(fired, Alert{
				JobName:    name,
				Reason:     reason,
				OccurredAt: time.Now(),
			})
		}
	}

	return fired
}
