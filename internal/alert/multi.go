package alert

import "errors"

// MultiHandler fans out a single alert to multiple handlers.
type MultiHandler struct {
	handlers []Handler
}

// NewMultiHandler creates a MultiHandler from the provided handlers.
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Handle calls every registered handler and collects errors.
// All handlers are called even if one returns an error.
func (m *MultiHandler) Handle(a Alert) error {
	var errs []error
	for _, h := range m.handlers {
		if err := h.Handle(a); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Add appends a handler to the MultiHandler at runtime.
func (m *MultiHandler) Add(h Handler) {
	m.handlers = append(m.handlers, h)
}

// Len returns the number of handlers registered with the MultiHandler.
func (m *MultiHandler) Len() int {
	return len(m.handlers)
}
