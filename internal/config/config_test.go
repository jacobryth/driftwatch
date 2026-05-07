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
	p := writeTmpConfig(t, "watch_targets:\n  - path: /tmp\n")
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("default poll_interval: got %v, want 30s", cfg.PollInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default log_level: got %q, want \"info\"", cfg.LogLevel)
	}
	if !cfg.Alerts.Log {
		t.Error("default alerts.log should be true")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	p := writeTmpConfig(t, `
poll_interval: 10s
log_level: debug
watch_targets:
  - path: /etc/nginx
    recurse: true
alerts:
  email:
    enabled: true
    host: mail.example.com
    port: 465
    from: a@b.com
    to: ["c@d.com"]
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("poll_interval: got %v, want 10s", cfg.PollInterval)
	}
	if !cfg.Alerts.Email.Enabled {
		t.Error("expected email enabled")
	}
	if cfg.Alerts.Email.Host != "mail.example.com" {
		t.Errorf("email host: got %q", cfg.Alerts.Email.Host)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/driftwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoWatchTargets(t *testing.T) {
	p := writeTmpConfig(t, "log_level: info\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error when watch_targets is empty")
	}
}

func TestLoad_InvalidPollInterval(t *testing.T) {
	p := writeTmpConfig(t, "poll_interval: -5s\nwatch_targets:\n  - path: /tmp\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for non-positive poll_interval")
	}
}

func TestLoad_EmailToSlice(t *testing.T) {
	p := writeTmpConfig(t, `
watch_targets:
  - path: /tmp
alerts:
  email:
    enabled: true
    to:
      - first@example.com
      - second@example.com
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Alerts.Email.To) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(cfg.Alerts.Email.To))
	}
}
