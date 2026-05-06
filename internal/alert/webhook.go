package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookHandler sends alert events to an HTTP endpoint as JSON payloads.
type WebhookHandler struct {
	endpoint string
	client   *http.Client
}

type webhookPayload struct {
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Path      string `json:"path"`
	EventType string `json:"event_type"`
	Message   string `json:"message"`
}

// NewWebhookHandler creates a WebhookHandler that posts to the given endpoint.
func NewWebhookHandler(endpoint string, timeout time.Duration) *WebhookHandler {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookHandler{
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
	}
}

// Handle serialises the event and POSTs it to the configured webhook endpoint.
func (w *WebhookHandler) Handle(a Alert) error {
	payload := webhookPayload{
		Timestamp: a.Timestamp.UTC().Format(time.RFC3339),
		Severity:  a.Severity,
		Path:      a.Event.Path,
		EventType: string(a.Event.Type),
		Message:   a.Message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post to %s: %w", w.endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.endpoint)
	}
	return nil
}
