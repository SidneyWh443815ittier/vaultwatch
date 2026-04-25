package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all vaultwatch runtime configuration.
type Config struct {
	Vault    VaultConfig   `yaml:"vault"`
	Monitor  MonitorConfig `yaml:"monitor"`
	Alerting AlertConfig   `yaml:"alerting"`
}

type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

type MonitorConfig struct {
	PollInterval    time.Duration `yaml:"poll_interval"`
	WarnThreshold   time.Duration `yaml:"warn_threshold"`
	CriticalThreshold time.Duration `yaml:"critical_threshold"`
}

type AlertConfig struct {
	Senders []SenderConfig `yaml:"senders"`
	AzureMonitor *AzureMonitorConfig `yaml:"azure_monitor,omitempty"`
}

type SenderConfig struct {
	Type   string            `yaml:"type"`
	Params map[string]string `yaml:"params"`
}

// AzureMonitorConfig holds configuration for the Azure Monitor sender.
type AzureMonitorConfig struct {
	WorkspaceID string `yaml:"workspace_id"`
	SharedKey   string `yaml:"shared_key"`
	LogType     string `yaml:"log_type"`
}

const (
	defaultPollInterval     = 60 * time.Second
	defaultWarnThreshold    = 72 * time.Hour
	defaultCriticalThreshold = 24 * time.Hour
)

// Load reads a YAML config file and applies environment variable overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		cfg.Vault.Token = token
	}
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		cfg.Vault.Address = addr
	}

	if cfg.Monitor.PollInterval == 0 {
		cfg.Monitor.PollInterval = defaultPollInterval
	}
	if cfg.Monitor.WarnThreshold == 0 {
		cfg.Monitor.WarnThreshold = defaultWarnThreshold
	}
	if cfg.Monitor.CriticalThreshold == 0 {
		cfg.Monitor.CriticalThreshold = defaultCriticalThreshold
	}

	if cfg.Vault.Address == "" {
		return nil, errors.New("config: vault.address is required")
	}

	return &cfg, nil
}
