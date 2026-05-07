package alert

import (
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/watcher"
)

func TestHandlerFunc_CallsFunction(t *testing.T) {
	called := false
	f := HandlerFunc(func(evt watcher.Event) error {
		called = true
		return nil
	})
	if err := f.Handle(makeEvent(watcher.EventModified, "/etc/x")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected function to be called")
	}
}

func TestHandlerFunc_PropagatesError(t *testing.T) {
	sentinel := errors.New("boom")
	f := HandlerFunc(func(_ watcher.Event) error { return sentinel })
	if err := f.Handle(makeEvent(watcher.EventModified, "/etc/x")); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestNopHandler_AlwaysNil(t *testing.T) {
	var h NopHandler
	for _, kind := range []watcher.EventKind{
		watcher.EventModified,
		watcher.EventDeleted,
		watcher.EventCreated,
	} {
		if err := h.Handle(makeEvent(kind, "/etc/x")); err != nil {
			t.Fatalf("NopHandler returned non-nil error for kind %v: %v", kind, err)
		}
	}
}
