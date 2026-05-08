// Package contractharness provides reusable runtime setup for contract tests.
package contractharness

import (
	"io"
	"log/slog"
	"testing"

	apphttp "github.com/recova-app/backend-v2/internal/app/http"
	"github.com/recova-app/backend-v2/internal/platform/config"
)

// BuildServer returns isolated HTTP runtime for contract tests.
func BuildServer(t testing.TB) *apphttp.Server {
	t.Helper()

	cfg := config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-contract-test",
			AppEnv:    "test",
			Port:      "3000",
			APIPrefix: "/api/v1",
		},
		Security: config.SecurityConfig{
			CORSOrigins:      []string{"http://localhost:5173"},
			RequestBodyLimit: "1mb",
		},
		Observability: config.ObservabilityConfig{
			RequestIDHeader:      "x-request-id",
			HealthCheckTimeoutMs: 2000,
		},
	}

	srv, err := apphttp.NewServer(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("build contract test server: %v", err)
	}

	return srv
}
