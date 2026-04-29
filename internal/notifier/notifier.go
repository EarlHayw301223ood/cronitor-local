package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultBaseURL = "https://cronitor.link"

// Event represents a job execution event sent to Cronitor.
type Event struct {
	JobName  string
	Series   string
	Status   string // "run", "complete", "fail"
	Message  string
	Duration float64
}

// Notifier sends ping events to the Cronitor API.
type Notifier struct {
	APIKey  string
	BaseURL string
	client  *http.Client
}

// New creates a Notifier with the given API key.
func New(apiKey string) *Notifier {
	return &Notifier{
		APIKey:  apiKey,
		BaseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type pingPayload struct {
	Status  string  `json:"status"`
	Message string  `json:"message,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Series  string  `json:"series,omitempty"`
}

// Ping sends an event ping for a monitored job.
func (n *Notifier) Ping(event Event) error {
	payload := pingPayload{
		Status:   event.Status,
		Message:  event.Message,
		Duration: event.Duration,
		Series:   event.Series,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifier: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", n.BaseURL, n.APIKey, event.JobName)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notifier: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("notifier: send ping: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notifier: unexpected status %d for job %q", resp.StatusCode, event.JobName)
	}
	return nil
}
