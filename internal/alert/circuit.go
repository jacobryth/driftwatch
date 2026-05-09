package alert

import (
	"fmt"
	"sync"
	"time"
)

// circuitState represents the state of the circuit breaker.
type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreakerHandler wraps a Handler and opens the circuit after a
// threshold of consecutive failures, preventing further calls until a
// cooldown period has elapsed.
type CircuitBreakerHandler struct {
	next      Handler
	threshold int
	cooldown  time.Duration
	nowFn     func() time.Time

	mu         sync.Mutex
	state      circuitState
	failures   int
	openedAt   time.Time
}

// NewCircuitBreakerHandler returns a CircuitBreakerHandler that trips open
// after threshold consecutive failures and resets after cooldown.
func NewCircuitBreakerHandler(next Handler, threshold int, cooldown time.Duration) *CircuitBreakerHandler {
	return &CircuitBreakerHandler{
		next:      next,
		threshold: threshold,
		cooldown:  cooldown,
		nowFn:     time.Now,
		state:     stateClosed,
	}
}

// Handle forwards the event to the wrapped handler unless the circuit is open.
func (c *CircuitBreakerHandler) Handle(evt Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case stateOpen:
		if c.nowFn().Sub(c.openedAt) < c.cooldown {
			return fmt.Errorf("circuit open: too many consecutive handler failures")
		}
		// Cooldown elapsed — allow one probe attempt.
		c.state = stateHalfOpen

	case stateClosed, stateHalfOpen:
		// proceed
	}

	err := c.next.Handle(evt)
	if err != nil {
		c.failures++
		if c.state == stateHalfOpen || c.failures >= c.threshold {
			c.state = stateOpen
			c.openedAt = c.nowFn()
		}
		return err
	}

	// Success: reset.
	c.failures = 0
	c.state = stateClosed
	return nil
}
