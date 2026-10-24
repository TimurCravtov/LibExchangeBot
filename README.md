# libexchangebot

Telegram bot written in Go using [`github.com/go-telegram/bot`](https://github.com/go-telegram/bot).

## Project Structure

```text
.
├── cmd/
│   └── libexchangebot/     # Application entrypoint (main package)
├── internal/
│   ├── bot/                # Telegram bot setup, middlewares, and handlers
│   └── config/             # Environment configuration loader
├── .env.example            # Example environment variables
├── go.mod                  # Go module definition
└── Makefile                # Common development tasks
```

## Getting Started

### Prerequisites

- Go 1.25+
- A Telegram Bot Token from [@BotFather](https://t.me/BotFather)

### Run

```sh
export TELEGRAM_BOT_TOKEN="your-token-here"
go run ./cmd/libexchangebot
```

On PowerShell:

```powershell
$env:TELEGRAM_BOT_TOKEN="your-token-here"
go run ./cmd/libexchangebot
```

### Build & Test

```sh
go vet ./...
go test -race ./...
go build -o bin/libexchangebot ./cmd/libexchangebot
```
