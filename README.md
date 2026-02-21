# Port Scanner Alert System

Network port monitoring tool built on top of Masscan. Scans targets, grabs banners, detects changes and sends alerts via Telegram and email.

## Документация на русском

https://docs.google.com/document/d/1utFbbpa8TUC48n_SeSUyEx7IoRJyFNX5pF_kDyZpf1U/edit?usp=sharing

## Features

- Masscan-based port scanning with banner grabbing
- Change detection: new, changed and closed ports
- Telegram and email notifications
- SQLite storage for scan history
- Scheduled periodic scanning
- Ansible playbook for automated deployment

## Requirements

- Go 1.24+
- Masscan (`apt install masscan`)
- libpcap (`apt install libpcap-dev`)
- Root privileges (required by Masscan)

## Project Structure

```
.
├── cmd/app/main.go              # entrypoint
├── config/config.yaml           # scan configuration
├── internal/
│   ├── config/                  # config loader
│   ├── diff/                    # scan result comparison
│   ├── logger/                  # zap logger setup
│   ├── model/                   # data models
│   ├── notifier/                # telegram and email senders
│   ├── report/                  # text and html report builders
│   ├── scanner/                 # masscan wrapper
│   ├── scheduler/               # periodic task runner
│   └── storage/sqlite/          # sqlite persistence
├── deploy/                      # ansible deployment
│   ├── inventory.ini
│   ├── playbook.yaml
│   ├── Makefile
│   └── templates/
└── .env                         # secrets (not committed)
```

## Quick Start

### 1. Clone and build

```bash
git clone https://github.com/Qwental/port-scanner-alert-system.git
cd port-scanner-alert-system
go build -o bin/scanner cmd/app/main.go
```

### 2. Configure

Edit `config/config.yaml`:

```yaml
project_name: "Port-scanner-alert-system"

masscan:
  rate: "1000"
  interface: ""
  ports: "22,53,80,443,1883,8080,9883"

targets:
  - "192.168.64.5"

database:
  path: "./scan_results.db"

scheduler:
  enabled: false
  interval: "30m"

telegram:
  enabled: true

smtp:
  enabled: false
```

### 3. Set up secrets

Create `.env` file in the project root:

```bash
TELEGRAM_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_CHAT_ID=123456789

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=you@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=you@gmail.com
SMTP_TO=recipient@example.com
```

For Gmail, generate an app password at https://myaccount.google.com/apppasswords

### 4. Run

```bash
sudo ./bin/scanner
```

Or with a custom config path:

```bash
sudo CONFIG_PATH=/etc/scanner/config.yaml ./bin/scanner
```

## Scheduled Mode

Set `scheduler.enabled: true` in config. The scanner will run repeatedly at the specified interval:

```yaml
scheduler:
  enabled: true
  interval: "30m"
```

Stop with `Ctrl+C` — the process handles SIGINT/SIGTERM gracefully.

## Notifications

### Telegram

1. Create a bot via [@BotFather](https://t.me/BotFather)
2. Get your chat ID via [@userinfobot](https://t.me/userinfobot)
3. Set `TELEGRAM_TOKEN` and `TELEGRAM_CHAT_ID` in `.env`
4. Set `telegram.enabled: true` in config

### Email (SMTP)

1. Set SMTP credentials in `.env`
2. Set `smtp.enabled: true` in config

Notifications are only sent when changes are detected (new, changed or closed ports).

## Deployment with Ansible

Deploy to a remote Ubuntu server with one command:

```bash
cd deploy
make deploy
```

This will:
- Install masscan and libpcap
- Create a system user and directories
- Copy the binary and config
- Set up a systemd service
- Start the scanner

See `deploy/` directory for inventory, playbook and templates.

### Prerequisites

- Ansible installed locally
- SSH access to the target server
- Cross-compilation toolchain for Linux (if building from macOS):

```bash
# macOS: install cross-compiler
brew install messense/macos-cross-toolchains/x86_64-unknown-linux-gnu
```

### Manual build for Linux

```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  CC=x86_64-linux-gnu-gcc \
  go build -ldflags="-s -w" -o bin/scanner-linux-amd64 cmd/app/main.go
```

## How It Works

1. Masscan scans the specified targets and ports with banner grabbing
2. Results are parsed from JSON output
3. Current results are compared with previous state from SQLite
4. New, changed and closed ports are identified
5. Results are saved to the database
6. If changes are found, notifications are sent via configured channels

## License
MIT
