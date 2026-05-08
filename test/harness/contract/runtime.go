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
	aimodule "github.com/recova-app/backend-v2/internal/modules/ai"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	communitymodule "github.com/recova-app/backend-v2/internal/modules/community"
	contentmodule "github.com/recova-app/backend-v2/internal/modules/content"
	educationmodule "github.com/recova-app/backend-v2/internal/modules/education"
	journalsmodule "github.com/recova-app/backend-v2/internal/modules/journals"
	routinemodule "github.com/recova-app/backend-v2/internal/modules/routine"
	usersmodule "github.com/recova-app/backend-v2/internal/modules/users"
	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/observability"
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
			RateLimit: config.RateLimitConfig{
				WindowMs: 60000,
				Max:      120,
				AuthMax:  10,
				AIMax:    20,
			},
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
	routineService := routinemodule.NewService(routinemodule.NewRepository(nil))
	journalsService := journalsmodule.NewService(journalsmodule.NewRepository(nil))
	communityService := communitymodule.NewService(communitymodule.NewRepository(nil))
	educationService := educationmodule.NewService(educationmodule.NewRepository(nil))
	contentService := contentmodule.NewService(contentmodule.NewRepository(nil))
	aiService := aimodule.NewService(aimodule.NewRepository(nil), &noopAIProvider{})

	srv, err := apphttp.NewServer(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)),
		apphttp.WithObservability(observability.NewRecorder()),
		apphttp.WithModuleDependencies(apphttp.ModuleDependencies{
			AuthService:      authService,
			UsersService:     usersService,
			RoutineService:   routineService,
			JournalsService:  journalsService,
			CommunityService: communityService,
			EducationService: educationService,
			ContentService:   contentService,
			AIService:        aiService,
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

type noopAIProvider struct{}

func (p *noopAIProvider) Generate(_ context.Context, _ aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	return aiplatform.GenerateResponse{}, errors.New("noop provider")
}
