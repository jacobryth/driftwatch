package alert

import (
	"errors"
	"testing"

	"github.com/yourusername/driftwatch/internal/watcher"
)

// captureHandler records the last event it received and the number of calls.
type captureHandler struct {
	calls  int
	last   watcher.Event
	retErr error
}

func (c *captureHandler) Handle(evt watcher.Event) error {
	c.calls++
	c.last = evt
	return c.retErr
}

func TestFilterHandler_NoPatterns_ForwardsAll(t *testing.T) {
	cap := &captureHandler{}
	f := NewFilterHandler(cap, nil)

	evt := watcher.Event{Path: "/etc/app/config.yaml"}
	if err := f.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.calls != 1 {
		t.Fatalf("expected 1 call, got %d", cap.calls)
	}
}

func TestFilterHandler_MatchingPattern_Forwards(t *testing.T) {
	cap := &captureHandler{}
	f := NewFilterHandler(cap, []string{"*.yaml"})

	evt := watcher.Event{Path: "/etc/app/config.yaml"}
	if err := f.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.calls != 1 {
		t.Fatalf("expected 1 call, got %d", cap.calls)
	}
}

func TestFilterHandler_NonMatchingPattern_Suppresses(t *testing.T) {
	cap := &captureHandler{}
	f := NewFilterHandler(cap, []string{"*.conf"})

	evt := watcher.Event{Path: "/etc/app/config.yaml"}
	if err := f.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", cap.calls)
	}
}

func TestFilterHandler_FullPathPattern_Matches(t *testing.T) {
	cap := &captureHandler{}
	f := NewFilterHandler(cap, []string{"/etc/app/*"})

	evt := watcher.Event{Path: "/etc/app/nginx.conf"}
	if err := f.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.calls != 1 {
		t.Fatalf("expected 1 call, got %d", cap.calls)
	}
}

func TestFilterHandler_PropagatesError(t *testing.T) {
	sentinel := errors.New("handler error")
	cap := &captureHandler{retErr: sentinel}
	f := NewFilterHandler(cap, []string{"*.yaml"})

	evt := watcher.Event{Path: "/etc/app/config.yaml"}
	if err := f.Handle(evt); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestFilterHandler_MultiplePatterns_MatchesFirst(t *testing.T) {
	cap := &captureHandler{}
	f := NewFilterHandler(cap, []string{"*.toml", "*.yaml", "*.json"})

	evt := watcher.Event{Path: "/srv/config.yaml"}
	if err := f.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.calls != 1 {
		t.Fatalf("expected 1 call, got %d", cap.calls)
	}
}
