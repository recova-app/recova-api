// Package e2eharness provides runtime assembly for end-to-end API flow tests.
package e2eharness

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
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
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/observability"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

const (
	validGoogleToken = "token-e2e-user-1"
)

// Runtime represents end-to-end test runtime dependencies.
type Runtime struct {
	Server *apphttp.Server
	DB     *database.Client
	Config config.Config
}

// NewRuntime constructs a DB-backed server with fake external dependencies for E2E tests.
func NewRuntime(t testing.TB) *Runtime {
	t.Helper()

	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	cfg := buildConfig(databaseURL)
	client, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})
	t.Cleanup(func() {
		databaseharness.ResetMigrations(t, databaseURL)
	})

	if err := seedReferenceContent(context.Background(), client); err != nil {
		t.Fatalf("seed reference content: %v", err)
	}

	authService := authmodule.NewService(
		authmodule.NewRepository(client.Gorm()),
		&fakeGoogleVerifier{},
		authmodule.NewTokenManager(cfg),
	)
	usersService := usersmodule.NewService(usersmodule.NewRepository(client.Gorm()), cfg.Application.AppEnv, cfg.Application.NodeEnv)
	routineService := routinemodule.NewService(routinemodule.NewRepository(client.Gorm()))
	journalsService := journalsmodule.NewService(journalsmodule.NewRepository(client.Gorm()))
	communityService := communitymodule.NewService(communitymodule.NewRepository(client.Gorm()))
	educationService := educationmodule.NewService(educationmodule.NewRepository(client.Gorm()))
	contentService := contentmodule.NewService(contentmodule.NewRepository(client.Gorm()))
	aiService := aimodule.NewService(aimodule.NewRepository(client.Gorm()), &fakeAIProvider{})

	recorder := observability.NewRecorder()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	srv, err := apphttp.NewServer(
		cfg,
		logger,
		apphttp.WithObservability(recorder),
		apphttp.WithReadinessChecks([]apphttp.ReadinessCheck{{
			Name:    "database",
			Mode:    apphttp.ReadinessModeRequired,
			Message: "Koneksi database tidak sehat",
			Probe:   client.Ping,
		}}),
		apphttp.WithModuleDependencies(apphttp.ModuleDependencies{
			AuthService:      authService,
			UsersService:     usersService,
			RoutineService:   routineService,
			JournalsService:  journalsService,
			CommunityService: communityService,
			EducationService: educationService,
			ContentService:   contentService,
			AIService:        aiService,
		}),
	)
	if err != nil {
		t.Fatalf("build e2e server: %v", err)
	}

	return &Runtime{Server: srv, DB: client, Config: cfg}
}

func buildConfig(databaseURL string) config.Config {
	return config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-e2e-test",
			AppEnv:    "local",
			NodeEnv:   "development",
			Port:      "3100",
			APIPrefix: "/api/v1",
			AppURL:    "http://localhost:3100",
			DocsURL:   "https://docs.recova.app",
		},
		Database: config.DatabaseConfig{
			URL:                databaseURL,
			DirectURL:          databaseURL,
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLifetimeSec: 300,
			SSLMode:            "disable",
		},
		Auth: config.AuthConfig{
			JWTSecret:     "e2e-jwt-secret-1234567890",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 7 * 24 * time.Hour,
			GoogleClient:  "e2e-google-client-id",
			Cookie: config.CookieConfig{
				Name:     "recova_refresh_e2e",
				Secure:   false,
				SameSite: "lax",
			},
		},
		Security: config.SecurityConfig{
			CORSOrigins:      []string{"http://localhost:5173"},
			RequestBodyLimit: "1mb",
			RateLimit: config.RateLimitConfig{
				WindowMs: 60000,
				Max:      1000,
				AuthMax:  1000,
				AIMax:    1000,
			},
		},
		Observability: config.ObservabilityConfig{
			RequestIDHeader:      "x-request-id",
			HealthCheckTimeoutMs: 3000,
		},
		Logger: config.LoggerConfig{
			Level:     "error",
			SlogLevel: slog.LevelError,
		},
	}
}

func seedReferenceContent(ctx context.Context, client *database.Client) error {
	if client == nil || client.Gorm() == nil {
		return errors.New("database client is not ready")
	}

	statements := []string{
		`INSERT INTO education_contents (id, title, description, url, thumbnail_url, category, is_active, published_at)
VALUES
  ('11111111-1111-1111-1111-111111111111','Memahami Trigger dan Rutinitas','Dasar mengenali pemicu harian dan membentuk respon yang lebih sehat.','https://recova.app/education/memahami-trigger-dan-rutinitas',NULL,'mindset',true,now()),
  ('22222222-2222-2222-2222-222222222222','Teknik Grounding 5-4-3-2-1','Latihan sederhana untuk kembali fokus saat dorongan muncul.','https://recova.app/education/teknik-grounding-5-4-3-2-1',NULL,'coping',true,now())
ON CONFLICT (id) DO NOTHING;`,
		`INSERT INTO daily_motivations (id, content, is_active, created_at)
VALUES
  ('33333333-3333-3333-3333-333333333333','Satu keputusan sehat hari ini tetap berarti besar.',true,now()),
  ('44444444-4444-4444-4444-444444444444','Kemajuan kecil yang konsisten lebih kuat dari niat sesaat.',true,now())
ON CONFLICT (content) DO NOTHING;`,
		`INSERT INTO daily_challenges (id, content, is_active, created_at)
VALUES
  ('55555555-5555-5555-5555-555555555555','Catat satu pemicu utama hari ini dan rencana responnya.',true,now()),
  ('66666666-6666-6666-6666-666666666666','Lakukan jeda 60 detik sebelum bereaksi saat dorongan muncul.',true,now())
ON CONFLICT (content) DO NOTHING;`,
	}

	for _, stmt := range statements {
		if err := client.Gorm().WithContext(ctx).Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
}

// ValidGoogleToken returns deterministic token accepted by fake verifier.
func ValidGoogleToken() string {
	return validGoogleToken
}

type fakeGoogleVerifier struct{}

func (v *fakeGoogleVerifier) Verify(_ context.Context, rawToken string, _ string) (authmodule.GoogleIdentity, error) {
	if strings.TrimSpace(rawToken) != validGoogleToken {
		return authmodule.GoogleIdentity{}, errors.New("invalid e2e token")
	}

	return authmodule.GoogleIdentity{
		GoogleID:    "google-e2e-user-1",
		Email:       "e2e-user-1@example.test",
		DisplayName: "E2E User",
	}, nil
}

type fakeAIProvider struct{}

func (p *fakeAIProvider) Generate(_ context.Context, req aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	if req.ForceJSON {
		return aiplatform.GenerateResponse{
			Provider: aiplatform.ProviderOpenAICompatible,
			Model:    "fake-e2e",
			Text:     `{"level":"Sedang","title":"Kamu sudah punya komitmen","level_description":"Ada dorongan yang masih perlu dikelola.","pattern_analysis":"Pemicu muncul saat lelah dan sendirian.","encouragement":"Ambil jeda napas 60 detik saat dorongan muncul."}`,
		}, nil
	}

	return aiplatform.GenerateResponse{
		Provider: aiplatform.ProviderOpenAICompatible,
		Model:    "fake-e2e",
		Text:     "Kamu tidak sendiri. Tarik napas dalam, lalu lakukan satu langkah kecil yang menenangkan sekarang.",
	}, nil
}
