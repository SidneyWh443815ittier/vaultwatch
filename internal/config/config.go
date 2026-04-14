package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultwatch configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines how and when alerts are sent.
type AlertsConfig struct {
	WarnBefore  time.Duration `yaml:"warn_before"`
	CritBefore  time.Duration `yaml:"crit_before"`
	SlackWebhook string       `yaml:"slack_webhook"`
	Email        EmailConfig  `yaml:"email"`
}

// EmailConfig holds SMTP settings for email alerts.
type EmailConfig struct {
	Enabled  bool     `yaml:"enabled"`
	SMTPHost string   `yaml:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
}

// MonitorConfig controls which secrets are monitored.
type MonitorConfig struct {
	Paths    []string      `yaml:"paths"`
	Interval time.Duration `yaml:"interval"`
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

	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}
	if c.Vault.Token == "" {
		if token := os.Getenv("VAULT_TOKEN"); token != "" {
			c.Vault.Token = token
		} else {
			return fmt.Errorf("vault.token is required (or set VAULT_TOKEN env var)")
		}
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Alerts.WarnBefore == 0 {
		c.Alerts.WarnBefore = 72 * time.Hour
	}
	if c.Alerts.CritBefore == 0 {
		c.Alerts.CritBefore = 24 * time.Hour
	}
	if c.Monitor.Interval == 0 {
		c.Monitor.Interval = 5 * time.Minute
	}
}
