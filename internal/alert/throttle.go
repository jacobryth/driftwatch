package alert

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/driftwatch/internal/watcher"
)

// ThrottleHandler suppresses repeated alerts for the same file within a
// sliding window, but always forwards the first occurrence.
type ThrottleHandler struct {
	next    Handler
	window  time.Duration
	mu      sync.Mutex
	lastAt  map[string]time.Time
	nowFunc func() time.Time
}

// NewThrottleHandler returns a Handler that forwards at most one alert per
// file path within window, delegating accepted events to next.
func NewThrottleHandler(next Handler, window time.Duration) *ThrottleHandler {
	return &ThrottleHandler{
		next:    next,
		window:  window,
		lastAt:  make(map[string]time.Time),
		nowFunc: time.Now,
	}
}

// Handle forwards the event if the file has not been alerted within the
// configured window. Suppressed events are silently dropped.
func (h *ThrottleHandler) Handle(evt watcher.Event) error {
	key := fmt.Sprintf("%s:%s", evt.Kind, evt.Path)

	h.mu.Lock()
	last, seen := h.lastAt[key]
	now := h.nowFunc()
	if seen && now.Sub(last) < h.window {
		h.mu.Unlock()
		return nil
	}
	h.lastAt[key] = now
	h.mu.Unlock()

	return h.next.Handle(evt)
}
