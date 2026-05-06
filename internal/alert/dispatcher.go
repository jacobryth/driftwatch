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

// Run blocks, processing events until ctx is cancelled or the events
// channel is closed. It logs a message when it stops processing.
func (d *Dispatcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			d.logf("dispatcher stopping: context cancelled: %v", ctx.Err())
			return
		case e, ok := <-d.events:
			if !ok {
				d.logf("dispatcher stopping: events channel closed")
				return
			}
			d.dispatch(e)
		}
	}
}

func (d *Dispatcher) dispatch(e watcher.Event) {
	a := NewAlert(e)
	if err := d.handler.Handle(a); err != nil {
		d.logf("alert handler error: %v", err)
	}
}

// logf writes a formatted log message if the dispatcher has a logger configured.
func (d *Dispatcher) logf(format string, args ...any) {
	if d.logger != nil {
		d.logger.Printf(format, args...)
	}
}
