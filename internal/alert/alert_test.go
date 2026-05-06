package alert_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/watcher"
)

func makeEvent(kind watcher.EventKind, path string) watcher.Event {
	return watcher.Event{Kind: kind, Path: path}
}

func TestNewAlert_Severity(t *testing.T) {
	cases := []struct {
		kind watcher.EventKind
		want alert.Severity
	}{
		{watcher.EventModified, alert.SeverityWarning},
		{watcher.EventDeleted, alert.SeverityCritical},
		{watcher.EventCreated, alert.SeverityInfo},
	}
	for _, tc := range cases {
		a := alert.NewAlert(makeEvent(tc.kind, "/etc/app/cfg"))
		if a.Severity != tc.want {
			t.Errorf("kind=%s: got severity %s, want %s", tc.kind, a.Severity, tc.want)
		}
	}
}

func TestLogHandler_Handle(t *testing.T) {
	var buf bytes.Buffer
	h := alert.NewLogHandler(&buf)
	e := makeEvent(watcher.EventModified, "/etc/app/config.yaml")
	a := alert.NewAlert(e)
	if err := h.Handle(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"severity=WARNING", "kind=modified", "/etc/app/config.yaml"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestMultiHandler_AllCalled(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	h1 := alert.NewLogHandler(&buf1)
	h2 := alert.NewLogHandler(&buf2)
	mh := alert.NewMultiHandler(h1, h2)

	a := alert.NewAlert(makeEvent(watcher.EventDeleted, "/etc/passwd"))
	if err := mh.Handle(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf1.Len() == 0 || buf2.Len() == 0 {
		t.Error("expected both handlers to receive the alert")
	}
}

func TestMultiHandler_CollectsErrors(t *testing.T) {
	failing := &failHandler{err: fmt.Errorf("write failed")}
	mh := alert.NewMultiHandler(failing, failing)
	a := alert.NewAlert(makeEvent(watcher.EventCreated, "/tmp/new"))
	err := mh.Handle(a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type failHandler struct{ err error }

func (f *failHandler) Handle(_ alert.Alert) error { return f.err }
