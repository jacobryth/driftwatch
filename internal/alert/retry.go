package alert

import (
	"fmt"
	"time"
)

// RetryHandler wraps a Handler and retries failed deliveries up to MaxAttempts
// times, waiting Delay between each attempt.
type RetryHandler struct {
	next        Handler
	maxAttempts int
	delay       time.Duration
	sleep       func(time.Duration)
}

// NewRetryHandler returns a RetryHandler that will retry next up to
// maxAttempts times (must be >= 1) with delay between attempts.
func NewRetryHandler(next Handler, maxAttempts int, delay time.Duration) *RetryHandler {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &RetryHandler{
		next:        next,
		maxAttempts: maxAttempts,
		delay:       delay,
		sleep:       time.Sleep,
	}
}

// Handle attempts to deliver the event, retrying on error.
func (r *RetryHandler) Handle(event Event) error {
	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		if err := r.next.Handle(event); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < r.maxAttempts {
			r.sleep(r.delay)
		}
	}
	return fmt.Errorf("all %d attempts failed: %w", r.maxAttempts, lastErr)
}
