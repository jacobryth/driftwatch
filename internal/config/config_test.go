package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTmpConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "driftwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTmpConfig: %v", err)
	}
	return p
}

func TestLoad_Defaults(t *testing.T) {
	p := writeTmpConfig(t, "watch:\n  - path: /etc/app\n")
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("default poll_interval: got %v, want 30s", cfg.PollInterval)
	}
	if cfg.Log.Format != "text" {
		t.Errorf("default log.format: got %q, want \"text\"", cfg.Log.Format)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	p := writeTmpConfig(t, `
poll_interval: 10s
watch:
  - path: /etc/nginx
    recurse: true
    excludes:
      - "*.bak"
log:
  level: debug
  format: json
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("poll_interval: got %v, want 10s", cfg.PollInterval)
	}
	if !cfg.Watch[0].Recurse {
		t.Error("expected recurse=true")
	}
	if cfg.Log.Format != "json" {
		t.Errorf("log.format: got %q, want \"json\"", cfg.Log.Format)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/driftwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoWatchTargets(t *testing.T) {
	p := writeTmpConfig(t, "poll_interval: 5s\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for empty watch list")
	}
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	p := writeTmpConfig(t, "watch:\n  - path: /etc\nlog:\n  format: xml\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for invalid log format")
	}
}

func TestLoad_EmptyWatchPath(t *testing.T) {
	p := writeTmpConfig(t, "watch:\n  - path: \"\"\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for empty watch path")
	}
}
