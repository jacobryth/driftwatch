package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

// Severity represents the urgency level of a drift alert.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Alert represents a single drift detection event.
type Alert struct {
	Timestamp time.Time
	Event     watcher.Event
	Severity  Severity
	Message   string
}

// Handler defines the interface for processing alerts.
type Handler interface {
	Handle(a Alert) error
}

// severityFor maps event kinds to severity levels.
func severityFor(e watcher.Event) Severity {
	switch e.Kind {
	case watcher.EventDeleted:
		return SeverityCritical
	case watcher.EventModified:
		return SeverityWarning
	case watcher.EventCreated:
		return SeverityInfo
	default:
		return SeverityInfo
	}
}

// NewAlert constructs an Alert from a watcher event.
func NewAlert(e watcher.Event) Alert {
	sev := severityFor(e)
	msg := fmt.Sprintf("[%s] %s: %s", sev, e.Kind, e.Path)
	return Alert{
		Timestamp: time.Now().UTC(),
		Event:     e,
		Severity:  sev,
		Message:   msg,
	}
}

// LogHandler writes alerts as structured lines to a writer.
type LogHandler struct {
	out io.Writer
}

// NewLogHandler creates a LogHandler writing to w.
// If w is nil, os.Stderr is used.
func NewLogHandler(w io.Writer) *LogHandler {
	if w == nil {
		w = os.Stderr
	}
	return &LogHandler{out: w}
}

// Handle writes the alert to the underlying writer.
func (h *LogHandler) Handle(a Alert) error {
	_, err := fmt.Fprintf(h.out, "time=%s severity=%s kind=%s path=%q\n",
		a.Timestamp.Format(time.RFC3339),
		a.Severity,
		a.Event.Kind,
		a.Event.Path,
	)
	return err
}
