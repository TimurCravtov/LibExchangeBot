package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	TelegramToken string
	LogLevel      slog.Level
	Debug         bool
}

// Load reads .env (overriding existing environment variables if present)
// and populates Config from environment variables.
func Load() (*Config, error) {
	if err := godotenv.Overload(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env file: %w", err)
	}

	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	debug := strings.EqualFold(os.Getenv("BOT_DEBUG"), "true") || os.Getenv("BOT_DEBUG") == "1"

	return &Config{
		TelegramToken: token,
		LogLevel:      level,
		Debug:         debug,
	}, nil
}

func parseLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
