package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackHandler sends drift alerts to a Slack incoming webhook URL.
type SlackHandler struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text        string       `json:"text"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

type slackAttachment struct {
	Color  string `json:"color"`
	Fields []slackField `json:"fields"`
}

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackHandler creates a SlackHandler that posts to the given Slack webhook URL.
func NewSlackHandler(webhookURL string) *SlackHandler {
	return &SlackHandler{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Handle sends the alert event to Slack.
func (s *SlackHandler) Handle(e Event) error {
	color := "#36a64f"
	if e.Severity == SeverityCritical {
		color = "#ff0000"
	} else if e.Severity == SeverityWarning {
		color = "#ffcc00"
	}

	payload := slackPayload{
		Text: fmt.Sprintf(":rotating_light: *DriftWatch Alert* — `%s`", e.Kind),
		Attachments: []slackAttachment{
			{
				Color: color,
				Fields: []slackField{
					{Title: "File", Value: e.Path, Short: true},
					{Title: "Severity", Value: string(e.Severity), Short: true},
					{Title: "Message", Value: e.Message, Short: false},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
