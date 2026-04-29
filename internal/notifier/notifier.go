package notifier

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// State represents the lifecycle event being reported.
type State string

const (
	StateRun      State = "run"
	StateComplete State = "complete"
	StateFail     State = "fail"
)

// Notifier sends ping events to a remote monitoring endpoint.
type Notifier struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// New creates a Notifier targeting baseURL with the given API key.
func New(baseURL, apiKey string) *Notifier {
	return &Notifier{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Ping sends a state notification for jobName.
func (n *Notifier) Ping(jobName string, state State) {
	url := fmt.Sprintf("%s/%s/%s?auth_key=%s", n.baseURL, jobName, state, n.apiKey)
	resp, err := n.client.Get(url)
	if err != nil {
		log.Printf("notifier: ping failed for %q (%s): %v", jobName, state, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("notifier: server returned %d for %q (%s)", resp.StatusCode, jobName, state)
	}
}
