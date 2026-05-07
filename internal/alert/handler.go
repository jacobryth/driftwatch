package alert

import "github.com/example/driftwatch/internal/watcher"

// Handler processes a single drift event. Implementations must be safe for
// concurrent use unless documented otherwise.
type Handler interface {
	Handle(evt watcher.Event) error
}

// HandlerFunc is a function that satisfies the Handler interface, analogous
// to http.HandlerFunc.
type HandlerFunc func(evt watcher.Event) error

// Handle calls f(evt).
func (f HandlerFunc) Handle(evt watcher.Event) error {
	return f(evt)
}

// NopHandler is a Handler that silently discards every event. Useful as a
// sentinel / default when no real handler is configured.
type NopHandler struct{}

// Handle discards the event and returns nil.
func (NopHandler) Handle(_ watcher.Event) error { return nil }
