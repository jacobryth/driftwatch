package alert

import (
	"math/rand"
	"sync"
	"time"
)

// SamplingHandler forwards only a statistical sample of events to the next
// handler. Events are sampled at the configured rate (0.0–1.0). A rate of 1.0
// forwards every event; 0.5 forwards approximately half.
type SamplingHandler struct {
	next Handler
	rate float64
	mu   sync.Mutex
	rng  *rand.Rand
}

// NewSamplingHandler returns a Handler that probabilistically forwards events
// to next at the given sample rate. rate must be between 0.0 and 1.0.
func NewSamplingHandler(next Handler, rate float64) *SamplingHandler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &SamplingHandler{
		next: next,
		rate: rate,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Handle forwards the event to the next handler if it passes the sample check.
func (s *SamplingHandler) Handle(event Event) error {
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()

	if v >= s.rate {
		return nil
	}
	return s.next.Handle(event)
}
