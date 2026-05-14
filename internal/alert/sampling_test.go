package alert

import (
	"errors"
	"sync/atomic"
	"testing"
)

func TestSamplingHandler_RateZero_DropsAll(t *testing.T) {
	var called int32
	next := HandlerFunc(func(e Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	h := NewSamplingHandler(next, 0.0)
	ev := makeEvent("modified", "/etc/app/cfg")

	for i := 0; i < 100; i++ {
		if err := h.Handle(ev); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if called != 0 {
		t.Errorf("expected 0 calls with rate=0, got %d", called)
	}
}

func TestSamplingHandler_RateOne_ForwardsAll(t *testing.T) {
	var called int32
	next := HandlerFunc(func(e Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	h := NewSamplingHandler(next, 1.0)
	ev := makeEvent("modified", "/etc/app/cfg")

	const n = 50
	for i := 0; i < n; i++ {
		if err := h.Handle(ev); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if int(called) != n {
		t.Errorf("expected %d calls with rate=1, got %d", n, called)
	}
}

func TestSamplingHandler_RateHalf_Approximate(t *testing.T) {
	var called int32
	next := HandlerFunc(func(e Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	h := NewSamplingHandler(next, 0.5)
	ev := makeEvent("modified", "/etc/app/cfg")

	const n = 10_000
	for i := 0; i < n; i++ {
		_ = h.Handle(ev)
	}

	got := int(called)
	// Allow ±10% tolerance around 50 %
	if got < n*40/100 || got > n*60/100 {
		t.Errorf("rate=0.5 over %d events: got %d forwarded, want ~%d", n, got, n/2)
	}
}

func TestSamplingHandler_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	next := HandlerFunc(func(e Event) error { return sentinel })

	h := NewSamplingHandler(next, 1.0)
	ev := makeEvent("modified", "/etc/app/cfg")

	if err := h.Handle(ev); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestSamplingHandler_ClampsBelowZero(t *testing.T) {
	var called int32
	next := HandlerFunc(func(e Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	h := NewSamplingHandler(next, -5.0)
	ev := makeEvent("modified", "/etc/app/cfg")
	_ = h.Handle(ev)

	if called != 0 {
		t.Errorf("negative rate should clamp to 0 and drop all events")
	}
}
