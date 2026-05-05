package watcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// FileState holds the last known state of a watched file.
type FileState struct {
	Path    string
	Checksum string
	ModTime time.Time
}

// ChangeEvent describes a detected change on a file.
type ChangeEvent struct {
	Path     string
	OldState *FileState
	NewState *FileState
	Kind     ChangeKind
}

// ChangeKind categorises the type of change detected.
type ChangeKind string

const (
	ChangeModified ChangeKind = "modified"
	ChangeDeleted  ChangeKind = "deleted"
	ChangeCreated  ChangeKind = "created"
)

// Watcher monitors a set of file paths for unexpected changes.
type Watcher struct {
	mu       sync.Mutex
	states   map[string]*FileState
	paths    []string
	Interval time.Duration
	Events   chan ChangeEvent
}

// New creates a Watcher for the given paths.
func New(paths []string, interval time.Duration) *Watcher {
	return &Watcher{
		states:   make(map[string]*FileState),
		paths:    paths,
		Interval: interval,
		Events:   make(chan ChangeEvent, 64),
	}
}

// Snapshot initialises the baseline state for all watched paths.
func (w *Watcher) Snapshot() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range w.paths {
		state, err := statFile(p)
		if err != nil {
			return fmt.Errorf("snapshot %s: %w", p, err)
		}
		w.states[p] = state
	}
	return nil
}

// Poll checks each watched path against the baseline and emits events.
func (w *Watcher) Poll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range w.paths {
		current, err := statFile(p)
		old := w.states[p]
		switch {
		case err != nil && os.IsNotExist(err):
			if old != nil {
				w.Events <- ChangeEvent{Path: p, OldState: old, Kind: ChangeDeleted}
				w.states[p] = nil
			}
		case err == nil && old == nil:
			w.Events <- ChangeEvent{Path: p, NewState: current, Kind: ChangeCreated}
			w.states[p] = current
		case err == nil && old != nil && current.Checksum != old.Checksum:
			w.Events <- ChangeEvent{Path: p, OldState: old, NewState: current, Kind: ChangeModified}
			w.states[p] = current
		}
	}
}

func statFile(path string) (*FileState, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return &FileState{
		Path:    path,
		Checksum: fmt.Sprintf("%x", h.Sum(nil)),
		ModTime: info.ModTime(),
	}, nil
}
