package alert

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

type countingHandler struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (c *countingHandler) Handle(_ watcher.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return c.err
}

func (c *countingHandler) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func TestRateLimitHandler_AllowsUpToMax(t *testing.T) {
	inner := &countingHandler{}
	h := NewRateLimitHandler(inner, 3, time.Minute)
	e := makeEvent(watcher.EventModified, "/etc/app.conf")

	for i := 0; i < 3; i++ {
		if err := h.Handle(e); err != nil {
			t.Fatalf("expected no error on call %d, got %v", i+1, err)
		}
	}
	if inner.count() != 3 {
		t.Fatalf("expected 3 forwarded events, got %d", inner.count())
	}
}

func TestRateLimitHandler_DropsOverLimit(t *testing.T) {
	inner := &countingHandler{}
	h := NewRateLimitHandler(inner, 2, time.Minute)
	e := makeEvent(watcher.EventModified, "/etc/app.conf")

	_ = h.Handle(e)
	_ = h.Handle(e)
	err := h.Handle(e)
	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}
	if inner.count() != 2 {
		t.Fatalf("expected 2 forwarded events, got %d", inner.count())
	}
}

func TestRateLimitHandler_ResetsAfterWindow(t *testing.T) {
	inner := &countingHandler{}
	current := time.Now()
	h := NewRateLimitHandler(inner, 1, time.Second)
	h.now = func() time.Time { return current }
	e := makeEvent(watcher.EventModified, "/etc/app.conf")

	if err := h.Handle(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := h.Handle(e); err == nil {
		t.Fatal("expected rate limit error before window reset")
	}

	// Advance past the window.
	current = current.Add(2 * time.Second)
	if err := h.Handle(e); err != nil {
		t.Fatalf("expected event after window reset, got: %v", err)
	}
	if inner.count() != 2 {
		t.Fatalf("expected 2 forwarded events, got %d", inner.count())
	}
}

func TestRateLimitHandler_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := &countingHandler{err: sentinel}
	h := NewRateLimitHandler(inner, 5, time.Minute)
	e := makeEvent(watcher.EventModified, "/etc/app.conf")

	if err := h.Handle(e); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
