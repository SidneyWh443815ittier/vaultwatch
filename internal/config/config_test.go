package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
vault:
  address: "https://vault.example.com"
  token: "s.testtoken"
monitor:
  paths:
    - secret/myapp
  interval: 10m
alerts:
  warn_before: 48h
  crit_before: 12h
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("unexpected vault address: %s", cfg.Vault.Address)
	}
	if cfg.Alerts.WarnBefore != 48*time.Hour {
		t.Errorf("unexpected warn_before: %v", cfg.Alerts.WarnBefore)
	}
	if cfg.Monitor.Interval != 10*time.Minute {
		t.Errorf("unexpected interval: %v", cfg.Monitor.Interval)
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, `
vault:
  address: "https://vault.example.com"
  token: "s.testtoken"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Alerts.WarnBefore != 72*time.Hour {
		t.Errorf("expected default warn_before 72h, got %v", cfg.Alerts.WarnBefore)
	}
	if cfg.Alerts.CritBefore != 24*time.Hour {
		t.Errorf("expected default crit_before 24h, got %v", cfg.Alerts.CritBefore)
	}
	if cfg.Monitor.Interval != 5*time.Minute {
		t.Errorf("expected default interval 5m, got %v", cfg.Monitor.Interval)
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	path := writeTempConfig(t, `
vault:
  token: "s.testtoken"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault address")
	}
}

func TestLoad_TokenFromEnv(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "s.envtoken")
	path := writeTempConfig(t, `
vault:
  address: "https://vault.example.com"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Vault.Token != "s.envtoken" {
		t.Errorf("expected token from env, got %q", cfg.Vault.Token)
	}
}
