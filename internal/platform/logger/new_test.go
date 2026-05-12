package logger

import (
	"log/slog"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

func TestNew_ReturnsLogger(t *testing.T) {
	cfg := config.Config{
		Logger: config.LoggerConfig{
			Level:     "info",
			SlogLevel: slog.LevelInfo,
		},
	}
	if New(cfg) == nil {
		t.Fatal("expected non-nil logger")
	}
}
