package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyHandler sends alerts to PagerDuty via the Events API v2.
type PagerDutyHandler struct {
	integrationKey string
	client         *http.Client
	eventsURL      string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
	CustomDetails map[string]string `json:"custom_details,omitempty"`
}

// NewPagerDutyHandler creates a PagerDutyHandler with the given integration key.
func NewPagerDutyHandler(integrationKey string) *PagerDutyHandler {
	return &PagerDutyHandler{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 10 * time.Second},
		eventsURL:      pagerDutyEventsURL,
	}
}

// Handle sends the alert event to PagerDuty.
func (p *PagerDutyHandler) Handle(a Alert) error {
	sev := "warning"
	if a.Severity == SeverityCritical {
		sev = "critical"
	}

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:   fmt.Sprintf("[driftwatch] %s: %s", a.Event.Kind, a.Event.Path),
			Source:    "driftwatch",
			Severity:  sev,
			Timestamp: a.Event.DetectedAt.UTC().Format(time.RFC3339),
			CustomDetails: map[string]string{
				"path": a.Event.Path,
				"kind": string(a.Event.Kind),
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := p.client.Post(p.eventsURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
