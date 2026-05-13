// Package logger provides the structured logger constructor for runtime bootstrap.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

const redactedSecretValue = "[REDACTED_SECRET]"

// New builds a structured logger for the current runtime environment.
func New(cfg config.Config) *slog.Logger {
	return NewWithWriter(cfg, os.Stdout)
}

// NewWithWriter builds a structured JSON logger with configurable output.
func NewWithWriter(cfg config.Config, output io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:       cfg.Logger.SlogLevel,
		ReplaceAttr: redactAttr,
	})

	return slog.New(handler)
}

func redactAttr(_ []string, attr slog.Attr) slog.Attr {
	if isSensitiveKey(attr.Key) {
		return slog.String(attr.Key, redactedSecretValue)
	}

	return attr
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}

	sensitiveHints := []string{
		"password",
		"secret",
		"token",
		"authorization",
		"api_key",
		"apikey",
		"cookie",
		"set-cookie",
		"prompt",
		"journal",
	}

	for _, hint := range sensitiveHints {
		if strings.Contains(normalized, hint) {
			return true
		}
	}

	return false
}
