package alert

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

// RateLimitHandler wraps a Handler and enforces a maximum number of alerts
// per window duration across all events. Once the limit is reached, events
// are dropped until the window resets.
type RateLimitHandler struct {
	inner   Handler
	max     int
	window  time.Duration
	now     func() time.Time

	mu       sync.Mutex
	count    int
	windowAt time.Time
}

// NewRateLimitHandler returns a Handler that forwards at most max events per
// window to inner. Subsequent events within the same window are silently
// dropped.
func NewRateLimitHandler(inner Handler, max int, window time.Duration) *RateLimitHandler {
	return &RateLimitHandler{
		inner:  inner,
		max:    max,
		window: window,
		now:    time.Now,
	}
}

func (r *RateLimitHandler) Handle(e watcher.Event) error {
	r.mu.Lock()
	now := r.now()
	if now.After(r.windowAt.Add(r.window)) {
		r.count = 0
		r.windowAt = now
	}
	if r.count >= r.max {
		r.mu.Unlock()
		return fmt.Errorf("rate limit exceeded: dropping event for %s", e.Path)
	}
	r.count++
	r.mu.Unlock()
	return r.inner.Handle(e)
}
