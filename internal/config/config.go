package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// WatchTarget describes a single directory or file to monitor.
type WatchTarget struct {
	Path        string   `yaml:"path"`
	Glob        string   `yaml:"glob"`
	Exclude     []string `yaml:"exclude"`
	Recursive   bool     `yaml:"recursive"`
}

// WebhookConfig holds optional webhook alerting settings.
type WebhookConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Endpoint string        `yaml:"endpoint"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Config is the top-level configuration for driftwatch.
type Config struct {
	PollInterval time.Duration  `yaml:"poll_interval"`
	LogLevel     string         `yaml:"log_level"`
	WatchTargets []WatchTarget  `yaml:"watch_targets"`
	Webhook      WebhookConfig  `yaml:"webhook"`
}

func defaults() Config {
	return Config{
		PollInterval: 30 * time.Second,
		LogLevel:     "info",
		Webhook: WebhookConfig{
			Timeout: 10 * time.Second,
		},
	}
}

// Load reads a YAML config file from path and applies defaults.
func Load(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.WatchTargets) == 0 {
		return nil, errors.New("config: at least one watch_target must be defined")
	}

	if cfg.Webhook.Enabled && cfg.Webhook.Endpoint == "" {
		return nil, errors.New("config: webhook.endpoint must be set when webhook is enabled")
	}

	if cfg.Webhook.Timeout == 0 {
		cfg.Webhook.Timeout = 10 * time.Second
	}

	return &cfg, nil
}
