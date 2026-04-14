# vaultwatch

A CLI tool that monitors HashiCorp Vault secret expiration and sends configurable alerts before leases expire.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Set your Vault address and token, then run vaultwatch with a config file:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

vaultwatch --config config.yaml
```

Example `config.yaml`:

```yaml
alert_threshold: 72h
notify:
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
secrets:
  - path: secret/data/db-credentials
  - path: secret/data/api-keys
```

vaultwatch will poll Vault at the configured interval and send alerts when any secret lease falls within the threshold window.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.yaml` | Path to config file |
| `--interval` | `1h` | Poll interval |
| `--dry-run` | `false` | Print alerts without sending |

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)