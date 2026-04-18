package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type AlertSender struct {
	Type       string `yaml:"type"`
	WebhookURL string `yaml:"webhook_url,omitempty"`
	APIKey     string `yaml:"api_key,omitempty"`
	RoutingKey string `yaml:"routing_key,omitempty"`
	Team       string `yaml:"team,omitempty"`
	FromEmail  string `yaml:"from_email,omitempty"`
	ToEmail    string `yaml:"to_email,omitempty"`
	SMTPHost   string `yaml:"smtp_host,omitempty"`
	SMTPPort   int    `yaml:"smtp_port,omitempty"`
}

type Thresholds struct {
	Warning  time.Duration `yaml:"warning"`
	Critical time.Duration `yaml:"critical"`
}

type Config struct {
	VaultAddress string      `yaml:"vault_address"`
	VaultToken   string      `yaml:"vault_token"`
	PollInterval time.Duration `yaml:"poll_interval"`
	Thresholds   Thresholds  `yaml:"thresholds"`
	Senders      []AlertSender `yaml:"senders"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}
	if cfg.VaultAddress == "" {
		return nil, errors.New("config: vault_address is required")
	}
	if cfg.VaultToken == "" {
		if tok := os.Getenv("VAULT_TOKEN"); tok != "" {
			cfg.VaultToken = tok
		}
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 60 * time.Second
	}
	if cfg.Thresholds.Warning == 0 {
		cfg.Thresholds.Warning = 72 * time.Hour
	}
	if cfg.Thresholds.Critical == 0 {
		cfg.Thresholds.Critical = 24 * time.Hour
	}
	return &cfg, nil
}

// SenderFactory builds Sender instances from config entries.
func (c *Config) SenderTypes() []string {
	types := make([]string, 0, len(c.Senders))
	for _, s := range c.Senders {
		types = append(types, s.Type)
	}
	return types
}
