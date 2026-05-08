// Package logger provides the structured logger constructor for runtime bootstrap.
package logger

import (
	"log/slog"
	"os"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

// New builds a structured logger for the current runtime environment.
func New(cfg config.AppConfig) *slog.Logger {
	if cfg.AppEnv == "local" {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
