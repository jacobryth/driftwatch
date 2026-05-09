package alert

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type collectHandler struct {
	mu     sync.Mutex
	events []Event
	errOn  int // return error on this call index (-1 = never)
	calls  int
}

func (c *collectHandler) Handle(e Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
	idx := c.calls
	c.calls++
	if c.errOn >= 0 && idx == c.errOn {
		return errors.New("handler error")
	}
	return nil
}

func (c *collectHandler) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.events)
}

func newBufEvent(path string) Event {
	return makeEvent(path, "modified")
}

func TestBufferHandler_FlushOnMaxSize(t *testing.T) {
	col := &collectHandler{errOn: -1}
	h := NewBufferHandler(col, 3, 10*time.Second)
	defer h.Stop() //nolint:errcheck

	_ = h.Handle(newBufEvent("/etc/a"))
	_ = h.Handle(newBufEvent("/etc/b"))
	if col.count() != 0 {
		t.Fatal("expected no flush before maxSize reached")
	}
	_ = h.Handle(newBufEvent("/etc/c"))
	if col.count() != 3 {
		t.Fatalf("expected 3 events after flush, got %d", col.count())
	}
}

func TestBufferHandler_FlushOnInterval(t *testing.T) {
	col := &collectHandler{errOn: -1}
	h := NewBufferHandler(col, 100, 50*time.Millisecond)
	defer h.Stop() //nolint:errcheck

	_ = h.Handle(newBufEvent("/etc/x"))
	time.Sleep(120 * time.Millisecond)
	if col.count() != 1 {
		t.Fatalf("expected 1 event after interval flush, got %d", col.count())
	}
}

func TestBufferHandler_StopFlushesRemaining(t *testing.T) {
	col := &collectHandler{errOn: -1}
	h := NewBufferHandler(col, 100, 10*time.Second)

	_ = h.Handle(newBufEvent("/etc/y"))
	_ = h.Handle(newBufEvent("/etc/z"))
	if err := h.Stop(); err != nil {
		t.Fatalf("unexpected error from Stop: %v", err)
	}
	if col.count() != 2 {
		t.Fatalf("expected 2 events after Stop, got %d", col.count())
	}
}

func TestBufferHandler_PropagatesHandlerError(t *testing.T) {
	col := &collectHandler{errOn: 0}
	h := NewBufferHandler(col, 1, 10*time.Second)
	defer h.Stop() //nolint:errcheck

	err := h.Handle(newBufEvent("/etc/fail"))
	if err == nil {
		t.Fatal("expected error from downstream handler")
	}
}

func TestBufferHandler_EmptyFlushIsNoop(t *testing.T) {
	col := &collectHandler{errOn: -1}
	h := NewBufferHandler(col, 5, 10*time.Second)
	if err := h.Stop(); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if col.count() != 0 {
		t.Fatal("expected no events on empty flush")
	}
}
