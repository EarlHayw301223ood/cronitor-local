# cronitor-local

Lightweight daemon that monitors local cron jobs and sends alerts on missed or failed executions.

---

## Installation

```bash
go install github.com/your-org/cronitor-local@latest
```

Or build from source:

```bash
git clone https://github.com/your-org/cronitor-local.git
cd cronitor-local
go build -o cronitor-local .
```

---

## Usage

Start the daemon by pointing it at your crontab:

```bash
cronitor-local --crontab /etc/crontab --alert-email ops@example.com
```

You can also watch a specific user's crontab and set a tolerance window for missed jobs:

```bash
cronitor-local --user deploy --tolerance 5m --webhook https://hooks.example.com/alert
```

**Key flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--crontab` | Path to crontab file | `/etc/crontab` |
| `--tolerance` | Grace period before alerting on a missed job | `2m` |
| `--alert-email` | Email address to send failure alerts | — |
| `--webhook` | Webhook URL for alert notifications | — |
| `--interval` | How often the daemon checks job status | `30s` |

---

## How It Works

`cronitor-local` parses your crontab, tracks the expected execution schedule for each job, and alerts you when a job fails to run within the configured tolerance window or exits with a non-zero status code.

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

This project is licensed under the [MIT License](LICENSE).