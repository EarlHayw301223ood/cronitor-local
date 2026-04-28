package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job represents a single monitored cron job entry.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Command  string        `yaml:"command"`
	Timeout  time.Duration `yaml:"timeout"`
	Notify   []string      `yaml:"notify"`
}

// Config holds the full application configuration.
type Config struct {
	APIKey      string        `yaml:"api_key"`
	PingURL     string        `yaml:"ping_url"`
	GracePeriod time.Duration `yaml:"grace_period"`
	Jobs        []Job         `yaml:"jobs"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if job.Command == "" {
			return fmt.Errorf("job %q: command is required", job.Name)
		}
	}
	return nil
}
