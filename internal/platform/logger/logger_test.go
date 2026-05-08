// Package logger tests structured logging and redaction behavior.
package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

// TestNewWithWriter_UsesJSONFormat ensures logger uses structured JSON output.
func TestNewWithWriter_UsesJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.Config{
		Logger: config.LoggerConfig{
			Level:     "info",
			SlogLevel: slog.LevelInfo,
		},
	}

	log := NewWithWriter(cfg, &buf)
	log.Info("startup", "service", "recova")

	output := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(output, "{") {
		t.Fatalf("expected JSON log output, got: %s", output)
	}

	if !strings.Contains(output, `"msg":"startup"`) {
		t.Fatalf("expected log message in output, got: %s", output)
	}
}

// TestNewWithWriter_RedactsSensitiveFields ensures secret fields are not logged in raw form.
func TestNewWithWriter_RedactsSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.Config{
		Logger: config.LoggerConfig{
			Level:     "info",
			SlogLevel: slog.LevelInfo,
		},
	}

	log := NewWithWriter(cfg, &buf)
	log.Info("auth event",
		"jwt_secret", "top-secret",
		"api_key", "sk-123",
		"user", "user-1",
	)

	output := buf.String()
	if strings.Contains(output, "top-secret") {
		t.Fatalf("secret value leaked: %s", output)
	}

	if strings.Contains(output, "sk-123") {
		t.Fatalf("api key leaked: %s", output)
	}

	if !strings.Contains(output, redactedSecretValue) {
		t.Fatalf("expected redacted marker not found: %s", output)
	}
}
