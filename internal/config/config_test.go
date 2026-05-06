package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/config"
)

func writeTmpConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTmpConfig(t, "watch_targets:\n  - path: /etc/app\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("poll_interval: want 30s, got %v", cfg.PollInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("log_level: want info, got %s", cfg.LogLevel)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	path := writeTmpConfig(t, "poll_interval: 10s\nlog_level: debug\nwatch_targets:\n  - path: /tmp\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("poll_interval: want 10s, got %v", cfg.PollInterval)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("log_level: want debug, got %s", cfg.LogLevel)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoWatchTargets(t *testing.T) {
	path := writeTmpConfig(t, "poll_interval: 5s\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when no watch_targets defined")
	}
}

func TestLoad_WebhookMissingEndpoint(t *testing.T) {
	path := writeTmpConfig(t, "watch_targets:\n  - path: /etc\nwebhook:\n  enabled: true\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when webhook enabled but endpoint missing")
	}
}

func TestLoad_WebhookValid(t *testing.T) {
	path := writeTmpConfig(t, "watch_targets:\n  - path: /etc\nwebhook:\n  enabled: true\n  endpoint: http://localhost:9000/hook\n  timeout: 5s\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Webhook.Enabled {
		t.Error("expected webhook to be enabled")
	}
	if cfg.Webhook.Timeout != 5*time.Second {
		t.Errorf("webhook timeout: want 5s, got %v", cfg.Webhook.Timeout)
	}
}
