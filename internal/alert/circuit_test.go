package alert

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func newCBHandler(t *testing.T, threshold int, cooldown time.Duration) (*CircuitBreakerHandler, *atomic.Int32, *atomic.Bool) {
	t.Helper()
	var calls atomic.Int32
	var fail atomic.Bool
	h := HandlerFunc(func(evt Event) error {
		calls.Add(1)
		if fail.Load() {
			return errors.New("downstream error")
		}
		return nil
	})
	cb := NewCircuitBreakerHandler(h, threshold, cooldown)
	return cb, &calls, &fail
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	cb, calls, _ := newCBHandler(t, 3, time.Minute)
	evt := makeEvent("modified")
	if err := cb.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb, calls, fail := newCBHandler(t, 3, time.Minute)
	fail.Store(true)
	evt := makeEvent("modified")
	for i := 0; i < 3; i++ {
		_ = cb.Handle(evt)
	}
	// Circuit should now be open — next call must not reach the handler.
	err := cb.Handle(evt)
	if err == nil {
		t.Fatal("expected circuit-open error, got nil")
	}
	if calls.Load() != 3 {
		t.Fatalf("expected exactly 3 downstream calls, got %d", calls.Load())
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	cb, _, fail := newCBHandler(t, 2, 50*time.Millisecond)
	fail.Store(true)
	evt := makeEvent("modified")
	_ = cb.Handle(evt)
	_ = cb.Handle(evt) // trips open

	// Simulate cooldown elapsed.
	cb.nowFn = func() time.Time { return time.Now().Add(100 * time.Millisecond) }

	// Probe attempt succeeds.
	fail.Store(false)
	if err := cb.Handle(evt); err != nil {
		t.Fatalf("expected success after cooldown, got: %v", err)
	}
	if cb.state != stateClosed {
		t.Fatalf("expected circuit closed after successful probe")
	}
}

func TestCircuitBreaker_ReopensOnHalfOpenFailure(t *testing.T) {
	cb, _, fail := newCBHandler(t, 2, 50*time.Millisecond)
	fail.Store(true)
	evt := makeEvent("modified")
	_ = cb.Handle(evt)
	_ = cb.Handle(evt) // trips open

	cb.nowFn = func() time.Time { return time.Now().Add(100 * time.Millisecond) }

	// Probe still fails — should re-open.
	_ = cb.Handle(evt)
	if cb.state != stateOpen {
		t.Fatalf("expected circuit to re-open after failed probe")
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	cb, _, fail := newCBHandler(t, 3, time.Minute)
	fail.Store(true)
	evt := makeEvent("modified")
	_ = cb.Handle(evt)
	_ = cb.Handle(evt) // 2 failures, not yet tripped

	fail.Store(false)
	if err := cb.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.failures != 0 {
		t.Fatalf("expected failure counter reset to 0, got %d", cb.failures)
	}
}
