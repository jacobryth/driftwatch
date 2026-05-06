package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/watcher"
)

func writeTmp(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTmp: %v", err)
	}
	return p
}

func TestSnapshot_UnknownFile(t *testing.T) {
	w := watcher.New([]string{"/nonexistent/path/config.yaml"}, time.Second)
	if err := w.Snapshot(); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestPoll_DetectsModification(t *testing.T) {
	dir := t.TempDir()
	p := writeTmp(t, dir, "app.conf", "key=value")

	w := watcher.New([]string{p}, time.Second)
	if err := w.Snapshot(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	// Modify the file.
	writeTmp(t, dir, "app.conf", "key=changed")
	w.Poll()

	select {
	case ev := <-w.Events:
		if ev.Kind != watcher.ChangeModified {
			t.Fatalf("expected modified, got %s", ev.Kind)
		}
		if ev.Path != p {
			t.Fatalf("unexpected path %s", ev.Path)
		}
	default:
		t.Fatal("expected a change event, got none")
	}
}

func TestPoll_DetectsDeletion(t *testing.T) {
	dir := t.TempDir()
	p := writeTmp(t, dir, "app.conf", "key=value")

	w := watcher.New([]string{p}, time.Second)
	if err := w.Snapshot(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	os.Remove(p)
	w.Poll()

	select {
	case ev := <-w.Events:
		if ev.Kind != watcher.ChangeDeleted {
			t.Fatalf("expected deleted, got %s", ev.Kind)
		}
	default:
		t.Fatal("expected a deletion event, got none")
	}
}

func TestPoll_NoEventWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	p := writeTmp(t, dir, "app.conf", "key=value")

	w := watcher.New([]string{p}, time.Second)
	if err := w.Snapshot(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	w.Poll()

	select {
	case ev := <-w.Events:
		t.Fatalf("unexpected event: %+v", ev)
	default:
		// pass
	}
}

func TestPoll_DetectsMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	p1 := writeTmp(t, dir, "a.conf", "x=1")
	p2 := writeTmp(t, dir, "b.conf", "y=2")

	w := watcher.New([]string{p1, p2}, time.Second)
	if err := w.Snapshot(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	// Modify both files.
	writeTmp(t, dir, "a.conf", "x=changed")
	writeTmp(t, dir, "b.conf", "y=changed")
	w.Poll()

	seen := map[string]bool{}
	for i := 0; i < 2; i++ {
		select {
		case ev := <-w.Events:
			if ev.Kind != watcher.ChangeModified {
				t.Fatalf("expected modified, got %s", ev.Kind)
			}
			seen[ev.Path] = true
		default:
			t.Fatalf("expected event %d, got none", i+1)
		}
	}

	if !seen[p1] || !seen[p2] {
		t.Fatalf("did not receive events for all modified files: %v", seen)
	}
}
