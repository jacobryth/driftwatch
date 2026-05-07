package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level driftwatch configuration.
type Config struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	LogLevel     string        `yaml:"log_level"`
	WatchTargets []WatchTarget `yaml:"watch_targets"`
	Alerts       AlertsConfig  `yaml:"alerts"`
}

// WatchTarget describes a single file or directory to monitor.
type WatchTarget struct {
	Path    string   `yaml:"path"`
	Ignore  []string `yaml:"ignore"`
	Recurse bool     `yaml:"recurse"`
}

// AlertsConfig aggregates all notification back-end settings.
type AlertsConfig struct {
	Log       bool              `yaml:"log"`
	Webhook   WebhookConfig     `yaml:"webhook"`
	Slack     SlackConfig       `yaml:"slack"`
	PagerDuty PagerDutyConfig   `yaml:"pagerduty"`
	Email     EmailAlertConfig  `yaml:"email"`
}

// WebhookConfig holds generic webhook settings.
type WebhookConfig struct {
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

// SlackConfig holds Slack webhook settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Enabled    bool   `yaml:"enabled"`
}

// PagerDutyConfig holds PagerDuty routing-key settings.
type PagerDutyConfig struct {
	RoutingKey string `yaml:"routing_key"`
	Enabled    bool   `yaml:"enabled"`
}

// EmailAlertConfig mirrors alert.EmailConfig plus an Enabled flag.
type EmailAlertConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
}

func defaults() Config {
	return Config{
		PollInterval: 30 * time.Second,
		LogLevel:     "info",
		Alerts:       AlertsConfig{Log: true},
	}
}

// Load reads and validates a YAML config file at path.
func Load(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if len(cfg.WatchTargets) == 0 {
		return nil, errors.New("config must define at least one watch_target")
	}

	if cfg.PollInterval <= 0 {
		return nil, errors.New("poll_interval must be positive")
	}

	return &cfg, nil
}
