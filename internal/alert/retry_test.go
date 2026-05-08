package alert

import (
	"errors"
	"testing"
	"time"
)

type countingHandler struct {
	calls  int
	failN  int // fail the first N calls
	err    error
}

func (c *countingHandler) Handle(_ Event) error {
	c.calls++
	if c.calls <= c.failN {
		return c.err
	}
	return nil
}

func newRetryHandler(t *testing.T, h Handler, attempts int) *RetryHandler {
	t.Helper()
	rh := NewRetryHandler(h, attempts, 0)
	rh.sleep = func(time.Duration) {} // no-op sleep in tests
	return rh
}

func TestRetryHandler_SucceedsOnFirstAttempt(t *testing.T) {
	ch := &countingHandler{failN: 0, err: errors.New("boom")}
	rh := newRetryHandler(t, ch, 3)
	if err := rh.Handle(makeEvent("modified")); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ch.calls != 1 {
		t.Fatalf("expected 1 call, got %d", ch.calls)
	}
}

func TestRetryHandler_RetriesOnFailure(t *testing.T) {
	ch := &countingHandler{failN: 2, err: errors.New("transient")}
	rh := newRetryHandler(t, ch, 3)
	if err := rh.Handle(makeEvent("modified")); err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if ch.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", ch.calls)
	}
}

func TestRetryHandler_ExhaustsAttempts(t *testing.T) {
	sentinel := errors.New("permanent")
	ch := &countingHandler{failN: 10, err: sentinel}
	rh := newRetryHandler(t, ch, 3)
	err := rh.Handle(makeEvent("modified"))
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped sentinel, got %v", err)
	}
	if ch.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", ch.calls)
	}
}

func TestRetryHandler_MinAttemptsIsOne(t *testing.T) {
	ch := &countingHandler{failN: 10, err: errors.New("err")}
	rh := newRetryHandler(t, ch, 0) // 0 should be clamped to 1
	err := rh.Handle(makeEvent("modified"))
	if err == nil {
		t.Fatal("expected error")
	}
	if ch.calls != 1 {
		t.Fatalf("expected exactly 1 call, got %d", ch.calls)
	}
}

func TestRetryHandler_SleepsBeforeRetry(t *testing.T) {
	sleptCount := 0
	ch := &countingHandler{failN: 2, err: errors.New("err")}
	rh := NewRetryHandler(ch, 3, 10*time.Millisecond)
	rh.sleep = func(d time.Duration) { sleptCount++ }
	_ = rh.Handle(makeEvent("modified"))
	// Should sleep between attempt 1→2 and 2→3 (not after final success)
	if sleptCount != 2 {
		t.Fatalf("expected 2 sleeps, got %d", sleptCount)
	}
}
