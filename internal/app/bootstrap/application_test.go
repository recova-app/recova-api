package bootstrap

import (
	"io"
	"log/slog"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

func TestNewApplication_InvalidDatabaseURL_ReturnsError(t *testing.T) {
	cfg := config.Config{
		Database: config.DatabaseConfig{
			URL:                "postgresql://invalid:invalid@127.0.0.1:1/recova_db?sslmode=disable",
			MaxOpenConns:       2,
			MaxIdleConns:       1,
			ConnMaxLifetimeSec: 60,
		},
		Observability: config.ObservabilityConfig{
			HealthCheckTimeoutMs: 50,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := NewApplication(cfg, logger); err == nil {
		t.Fatal("expected error")
	}
}
