package alert

import (
	"sync"
	"time"
)

// DedupeHandler wraps another Handler and suppresses duplicate alerts for the
// same file path within a configurable TTL window.
type DedupeHandler struct {
	inner   Handler
	ttl     time.Duration
	mu      sync.Mutex
	seen    map[string]time.Time
	nowFunc func() time.Time
}

// NewDedupeHandler returns a DedupeHandler that forwards events to inner only
// when the same path has not been seen within ttl.
func NewDedupeHandler(inner Handler, ttl time.Duration) *DedupeHandler {
	return &DedupeHandler{
		inner:   inner,
		ttl:     ttl,
		seen:    make(map[string]time.Time),
		nowFunc: time.Now,
	}
}

// Handle forwards the event to the inner handler unless an identical path was
// already forwarded within the TTL window.
func (d *DedupeHandler) Handle(e Event) error {
	now := d.nowFunc()

	d.mu.Lock()
	last, exists := d.seen[e.Path]
	if exists && now.Sub(last) < d.ttl {
		d.mu.Unlock()
		return nil
	}
	d.seen[e.Path] = now
	d.mu.Unlock()

	return d.inner.Handle(e)
}

// Purge removes expired entries from the seen map. Call periodically to avoid
// unbounded memory growth in long-running daemons.
func (d *DedupeHandler) Purge() {
	now := d.nowFunc()
	d.mu.Lock()
	defer d.mu.Unlock()
	for path, t := range d.seen {
		if now.Sub(t) >= d.ttl {
			delete(d.seen, path)
		}
	}
}
