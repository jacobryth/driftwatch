package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch configuration.
type Config struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	Watch        []WatchTarget  `yaml:"watch"`
	Log          LogConfig      `yaml:"log"`
}

// WatchTarget describes a single directory or file to monitor.
type WatchTarget struct {
	Path     string   `yaml:"path"`
	Recurse  bool     `yaml:"recurse"`
	Excludes []string `yaml:"excludes"`
}

// LogConfig controls structured log output.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // "json" or "text"
}

// Load reads and validates a YAML config file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := &Config{
		PollInterval: 30 * time.Second,
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("config: poll_interval must be positive")
	}
	if len(c.Watch) == 0 {
		return fmt.Errorf("config: at least one watch target is required")
	}
	for i, t := range c.Watch {
		if t.Path == "" {
			return fmt.Errorf("config: watch[%d]: path must not be empty", i)
		}
	}
	switch c.Log.Format {
	case "json", "text":
	default:
		return fmt.Errorf("config: log.format must be \"json\" or \"text\", got %q", c.Log.Format)
	}
	return nil
}
