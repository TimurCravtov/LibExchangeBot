package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DotEnvOverridesExistingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Setenv("TELEGRAM_BOT_TOKEN", "old-token")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("BOT_DEBUG", "false")

	envContent := "TELEGRAM_BOT_TOKEN=dotenv-token\nLOG_LEVEL=debug\nBOT_DEBUG=true\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.TelegramToken != "dotenv-token" {
		t.Errorf("TelegramToken = %q, want %q", cfg.TelegramToken, "dotenv-token")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelDebug)
	}
	if !cfg.Debug {
		t.Errorf("Debug = false, want true")
	}
}
