package alert

import (
	"sync"
	"time"
)

// BufferHandler batches events and flushes them to the next handler either
// when the buffer reaches maxSize or when the flush interval elapses.
type BufferHandler struct {
	next      Handler
	maxSize   int
	interval  time.Duration
	mu        sync.Mutex
	buf       []Event
	stopCh    chan struct{}
	nowFunc   func() time.Time
}

// NewBufferHandler returns a BufferHandler that accumulates events and forwards
// them as a batch. Call Stop to release the background flush goroutine.
func NewBufferHandler(next Handler, maxSize int, interval time.Duration) *BufferHandler {
	h := &BufferHandler{
		next:     next,
		maxSize:  maxSize,
		interval: interval,
		buf:      make([]Event, 0, maxSize),
		stopCh:   make(chan struct{}),
		nowFunc:  time.Now,
	}
	go h.flushLoop()
	return h
}

// Handle appends the event to the internal buffer. If the buffer reaches
// maxSize it is flushed immediately.
func (h *BufferHandler) Handle(e Event) error {
	h.mu.Lock()
	h.buf = append(h.buf, e)
	ready := len(h.buf) >= h.maxSize
	h.mu.Unlock()

	if ready {
		return h.flush()
	}
	return nil
}

// Stop signals the background goroutine to exit and performs a final flush.
func (h *BufferHandler) Stop() error {
	close(h.stopCh)
	return h.flush()
}

func (h *BufferHandler) flush() error {
	h.mu.Lock()
	if len(h.buf) == 0 {
		h.mu.Unlock()
		return nil
	}
	batch := make([]Event, len(h.buf))
	copy(batch, h.buf)
	h.buf = h.buf[:0]
	h.mu.Unlock()

	var firstErr error
	for _, ev := range batch {
		if err := h.next.Handle(ev); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (h *BufferHandler) flushLoop() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = h.flush()
		case <-h.stopCh:
			return
		}
	}
}
