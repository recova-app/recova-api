// Package contractharness provides reusable runtime setup for contract tests.
package contractharness

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	apphttp "github.com/recova-app/backend-v2/internal/app/http"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	usersmodule "github.com/recova-app/backend-v2/internal/modules/users"
	"github.com/recova-app/backend-v2/internal/platform/config"
)

// BuildServer returns isolated HTTP runtime for contract tests.
func BuildServer(t testing.TB) *apphttp.Server {
	t.Helper()

	cfg := config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-contract-test",
			AppEnv:    "test",
			NodeEnv:   "test",
			Port:      "3000",
			APIPrefix: "/api/v1",
		},
		Auth: config.AuthConfig{
			JWTSecret:     "contract-test-jwt-secret-123456",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 24 * time.Hour,
			GoogleClient:  "contract-test-google-client-id",
			Cookie: config.CookieConfig{
				Name:     "recova_refresh_contract",
				Secure:   false,
				SameSite: "lax",
			},
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

	authService := authmodule.NewService(
		authmodule.NewRepository(nil),
		&noopGoogleVerifier{},
		authmodule.NewTokenManager(cfg),
	)
	usersService := usersmodule.NewService(usersmodule.NewRepository(nil), cfg.Application.AppEnv, cfg.Application.NodeEnv)

	srv, err := apphttp.NewServer(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), apphttp.WithModuleDependencies(apphttp.ModuleDependencies{
		AuthService:  authService,
		UsersService: usersService,
	}))
	if err != nil {
		t.Fatalf("build contract test server: %v", err)
	}

	return srv
}

type noopGoogleVerifier struct{}

func (v *noopGoogleVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("noop verifier")
}
