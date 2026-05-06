package alert

import (
	"context"
	"log"

	"github.com/example/driftwatch/internal/watcher"
)

// Dispatcher consumes events from a channel and dispatches alerts.
type Dispatcher struct {
	events  <-chan watcher.Event
	handler Handler
	logger  *log.Logger
}

// NewDispatcher creates a Dispatcher that reads from events and
// forwards resulting alerts to handler.
func NewDispatcher(events <-chan watcher.Event, handler Handler, logger *log.Logger) *Dispatcher {
	return &Dispatcher{
		events:  events,
		handler: handler,
		logger:  logger,
	}
}

// Run blocks, processing events until ctx is cancelled.
func (d *Dispatcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-d.events:
			if !ok {
				return
			}
			d.dispatch(e)
		}
	}
}

func (d *Dispatcher) dispatch(e watcher.Event) {
	a := NewAlert(e)
	if err := d.handler.Handle(a); err != nil {
		if d.logger != nil {
			d.logger.Printf("alert handler error: %v", err)
		}
	}
}
