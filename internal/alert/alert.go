package alert

import (
	"fmt"
	"log"
	"time"
)

// EventKind describes what happened to a watched file.
type EventKind int

const (
	EventModified EventKind = iota
	EventDeleted
	EventCreated
)

func (k EventKind) String() string {
	switch k {
	case EventModified:
		return "modified"
	case EventDeleted:
		return "deleted"
	case EventCreated:
		return "created"
	default:
		return "unknown"
	}
}

// Severity indicates how urgent an alert is.
type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event carries information about a single drift detection.
type Event struct {
	Path      string
	Kind      EventKind
	Severity  Severity
	Timestamp time.Time
	Message   string
}

// Handler is implemented by any alert destination.
type Handler interface {
	Handle(Event) error
}

// severityFor maps event kinds to severities.
func severityFor(k EventKind) Severity {
	if k == EventDeleted {
		return SeverityCritical
	}
	return SeverityWarning
}

// NewAlert constructs an Event with sensible defaults.
func NewAlert(path string, kind EventKind) Event {
	sev := severityFor(kind)
	return Event{
		Path:      path,
		Kind:      kind,
		Severity:  sev,
		Timestamp: time.Now().UTC(),
		Message:   fmt.Sprintf("file %s: %s", path, kind),
	}
}

// logHandler writes alerts to the standard logger.
type logHandler struct{ logger *log.Logger }

// NewLogHandler returns a Handler that logs each event.
func NewLogHandler(logger *log.Logger) Handler {
	return &logHandler{logger: logger}
}

func (h *logHandler) Handle(e Event) error {
	h.logger.Printf("[%s] %s", e.Severity, e.Message)
	return nil
}
