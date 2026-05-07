package alert

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingHandler struct {
	calls atomic.Int32
	err   error
}

func (c *countingHandler) Handle(_ Event) error {
	c.calls.Add(1)
	return c.err
}

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestDedupeHandler_ForwardsFirstEvent(t *testing.T) {
	inner := &countingHandler{}
	d := NewDedupeHandler(inner, 5*time.Minute)
	e := makeEvent("/etc/app/cfg.yaml", EventModified)

	if err := d.Handle(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestDedupeHandler_SuppressesDuplicate(t *testing.T) {
	inner := &countingHandler{}
	now := time.Now()
	d := NewDedupeHandler(inner, 5*time.Minute)
	d.nowFunc = fixedNow(now)
	e := makeEvent("/etc/app/cfg.yaml", EventModified)

	_ = d.Handle(e)
	_ = d.Handle(e)

	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call after duplicate, got %d", got)
	}
}

func TestDedupeHandler_ForwardsAfterTTLExpiry(t *testing.T) {
	inner := &countingHandler{}
	now := time.Now()
	ttl := 1 * time.Minute
	d := NewDedupeHandler(inner, ttl)
	e := makeEvent("/etc/app/cfg.yaml", EventModified)

	d.nowFunc = fixedNow(now)
	_ = d.Handle(e)

	d.nowFunc = fixedNow(now.Add(ttl + time.Second))
	_ = d.Handle(e)

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after TTL expiry, got %d", got)
	}
}

func TestDedupeHandler_DifferentPathsBothForwarded(t *testing.T) {
	inner := &countingHandler{}
	d := NewDedupeHandler(inner, 5*time.Minute)

	_ = d.Handle(makeEvent("/etc/app/a.yaml", EventModified))
	_ = d.Handle(makeEvent("/etc/app/b.yaml", EventModified))

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls for distinct paths, got %d", got)
	}
}

func TestDedupeHandler_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := &countingHandler{err: sentinel}
	d := NewDedupeHandler(inner, 5*time.Minute)

	err := d.Handle(makeEvent("/etc/app/cfg.yaml", EventModified))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDedupeHandler_Purge(t *testing.T) {
	inner := &countingHandler{}
	now := time.Now()
	ttl := 1 * time.Minute
	d := NewDedupeHandler(inner, ttl)
	e := makeEvent("/etc/app/cfg.yaml", EventModified)

	d.nowFunc = fixedNow(now)
	_ = d.Handle(e)

	// Advance time past TTL and purge.
	d.nowFunc = fixedNow(now.Add(ttl + time.Second))
	d.Purge()

	if len(d.seen) != 0 {
		t.Fatalf("expected seen map to be empty after purge, got %d entries", len(d.seen))
	}
}
