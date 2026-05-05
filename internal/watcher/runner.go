package watcher

import (
	"context"
	"log"
)

// Runner wraps a Watcher with a blocking run loop.
type Runner struct {
	w      *Watcher
	logger *log.Logger
}

// NewRunner creates a Runner using the provided Watcher and logger.
func NewRunner(w *Watcher, logger *log.Logger) *Runner {
	return &Runner{w: w, logger: logger}
}

// Run takes a baseline snapshot then polls on the configured interval until
// ctx is cancelled. Change events are logged as they arrive.
func (r *Runner) Run(ctx context.Context) error {
	if err := r.w.Snapshot(); err != nil {
		return err
	}
	r.logger.Printf("driftwatch: baseline snapshot taken for %d path(s)", len(r.w.paths))

	ticker := newTicker(r.w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Println("driftwatch: shutting down")
			return ctx.Err()
		case <-ticker.C:
			r.w.Poll()
		case ev := <-r.w.Events:
			r.logEvent(ev)
		}
	}
}

func (r *Runner) logEvent(ev ChangeEvent) {
	switch ev.Kind {
	case ChangeModified:
		r.logger.Printf("ALERT [modified] %s | old_checksum=%s new_checksum=%s",
			ev.Path, ev.OldState.Checksum, ev.NewState.Checksum)
	case ChangeDeleted:
		r.logger.Printf("ALERT [deleted]  %s | last_checksum=%s",
			ev.Path, ev.OldState.Checksum)
	case ChangeCreated:
		r.logger.Printf("ALERT [created]  %s | checksum=%s",
			ev.Path, ev.NewState.Checksum)
	}
}

// tickerIface allows tests to substitute a fake ticker (unexported helper).
type tickerIface interface {
	Stop()
	C() <-chan struct{}
}

import "time"

func newTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}
