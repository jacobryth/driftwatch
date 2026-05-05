# driftwatch

Lightweight daemon that detects and alerts on unexpected config file changes in production directories.

---

## Installation

```bash
go install github.com/yourorg/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a config file (`driftwatch.yaml`) to define which directories to monitor:

```yaml
watch:
  - path: /etc/nginx
    alert: slack
  - path: /etc/app/config
    alert: log

alert:
  slack_webhook: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  log_file: /var/log/driftwatch.log

interval: 30s
```

Start the daemon:

```bash
driftwatch --config driftwatch.yaml
```

When an unexpected change is detected, driftwatch logs the event and sends an alert:

```
[ALERT] 2024-01-15T10:32:01Z - File modified: /etc/nginx/nginx.conf (expected hash: a1b2c3, got: d4e5f6)
```

Run as a systemd service for production use — see [`docs/systemd.md`](docs/systemd.md) for setup instructions.

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `driftwatch.yaml` | Path to config file |
| `--dry-run` | `false` | Log changes without sending alerts |
| `--verbose` | `false` | Enable debug output |

---

## License

MIT © yourorg