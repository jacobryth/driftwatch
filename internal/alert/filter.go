package alert

import (
	"path/filepath"
	"strings"

	"github.com/yourusername/driftwatch/internal/watcher"
)

// FilterHandler wraps another Handler and only forwards events whose
// file paths match at least one of the provided glob patterns.
// If no patterns are provided every event is forwarded.
type FilterHandler struct {
	next     Handler
	patterns []string
}

// NewFilterHandler creates a FilterHandler that forwards events to next only
// when the event's file path matches one of patterns (filepath.Match syntax).
func NewFilterHandler(next Handler, patterns []string) *FilterHandler {
	return &FilterHandler{
		next:     next,
		patterns: patterns,
	}
}

// Handle checks whether the event's path matches any configured pattern and,
// if so, delegates to the wrapped handler.
func (f *FilterHandler) Handle(evt watcher.Event) error {
	if len(f.patterns) == 0 {
		return f.next.Handle(evt)
	}
	if f.matches(evt.Path) {
		return f.next.Handle(evt)
	}
	return nil
}

func (f *FilterHandler) matches(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range f.patterns {
		// Try matching against the full path first, then just the base name.
		if ok, _ := filepath.Match(pattern, path); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, base); ok {
			return true
		}
		// Support simple suffix matching (e.g. ".conf").
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(path, strings.TrimPrefix(pattern, "*")) {
			return true
		}
	}
	return false
}
