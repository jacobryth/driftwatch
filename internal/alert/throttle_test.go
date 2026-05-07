package alert

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

func newThrottleHandler(window time.Duration) (*ThrottleHandler, *int32) {
	var calls int32
	next := HandlerFunc(func(evt watcher.Event) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	h := NewThrottleHandler(next, window)
	return h, &calls
}

func TestThrottleHandler_ForwardsFirstEvent(t *testing.T) {
	h, calls := newThrottleHandler(5 * time.Minute)
	evt := makeEvent(watcher.EventModified, "/etc/app/cfg.yaml")
	if err := h.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", atomic.LoadInt32(calls))
	}
}

func TestThrottleHandler_SuppressesWithinWindow(t *testing.T) {
	h, calls := newThrottleHandler(5 * time.Minute)
	evt := makeEvent(watcher.EventModified, "/etc/app/cfg.yaml")

	_ = h.Handle(evt)
	_ = h.Handle(evt)
	_ = h.Handle(evt)

	if atomic.LoadInt32(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", atomic.LoadInt32(calls))
	}
}

func TestThrottleHandler_ForwardsAfterWindowExpiry(t *testing.T) {
	h, calls := newThrottleHandler(10 * time.Millisecond)
	evt := makeEvent(watcher.EventModified, "/etc/app/cfg.yaml")

	_ = h.Handle(evt)
	time.Sleep(20 * time.Millisecond)
	_ = h.Handle(evt)

	if atomic.LoadInt32(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", atomic.LoadInt32(calls))
	}
}

func TestThrottleHandler_DifferentPathsAreIndependent(t *testing.T) {
	h, calls := newThrottleHandler(5 * time.Minute)

	_ = h.Handle(makeEvent(watcher.EventModified, "/etc/app/a.yaml"))
	_ = h.Handle(makeEvent(watcher.EventModified, "/etc/app/b.yaml"))

	if atomic.LoadInt32(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", atomic.LoadInt32(calls))
	}
}

func TestThrottleHandler_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	next := HandlerFunc(func(_ watcher.Event) error { return sentinel })
	h := NewThrottleHandler(next, 5*time.Minute)

	if err := h.Handle(makeEvent(watcher.EventModified, "/etc/app/cfg.yaml")); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
