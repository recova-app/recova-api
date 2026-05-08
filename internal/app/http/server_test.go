// Package http tests baseline HTTP runtime behavior and middleware contract.
package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/observability"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
	"github.com/recova-app/backend-v2/test/fixtures"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
)

// TestNewServer_HealthRoutes_ReturnExpectedStatus ensures live/ready health routes are exposed.
func TestNewServer_HealthRoutes_ReturnExpectedStatus(t *testing.T) {
	srv := buildTestServer(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "live", path: "/health/live"},
		{name: "ready", path: "/health/ready"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, tc.path, nil, map[string]string{
				"Origin": "http://localhost:5173",
			})
			httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
			httpharness.RequireSuccessEnvelope(t, resp.JSON)
		})
	}
}

// TestNewServer_NotFound_UsesStandardErrorEnvelope ensures unknown routes return API-standard error payload.
func TestNewServer_NotFound_UsesStandardErrorEnvelope(t *testing.T) {
	srv := buildTestServer(t)

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/unknown", nil, map[string]string{
		"x-request-id": "req-notfound-1",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusNotFound)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "NOT_FOUND")

	errorPayload := resp.JSON["error"].(map[string]any)
	if errorPayload["requestId"] != "req-notfound-1" {
		t.Fatalf("expected request id propagation, got: %v", errorPayload["requestId"])
	}
}

// TestNewServer_RecoverMiddleware_MapsPanicToInternalError ensures panic is recovered and mapped safely.
func TestNewServer_RecoverMiddleware_MapsPanicToInternalError(t *testing.T) {
	srv := buildTestServer(t)
	srv.readinessChecks = []ReadinessCheck{
		{
			Name: "database",
			Mode: ReadinessModeRequired,
			Probe: func(_ context.Context) error {
				panic("boom")
			},
		},
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/ready", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusInternalServerError)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "INTERNAL_ERROR")
}

// TestNewServer_RequestIDAndSecurityHeaders ensures request id and security/CORS headers are present.
func TestNewServer_RequestIDAndSecurityHeaders(t *testing.T) {
	srv := buildTestServer(t)

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/live", nil, map[string]string{
		"Origin":       "http://localhost:5173",
		"x-request-id": "req-header-1",
	})

	if got := resp.Header.Get("X-Request-Id"); got != "req-header-1" {
		t.Fatalf("expected request id header echoed, got: %q", got)
	}

	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected helmet header nosniff, got: %q", got)
	}

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected cors allow origin header, got: %q", got)
	}
}

// TestNewServer_ReadinessFailure_ReturnsServiceUnavailable ensures failed required dependency returns 503.
func TestNewServer_ReadinessFailure_ReturnsServiceUnavailable(t *testing.T) {
	srv := buildTestServer(t)
	srv.readinessChecks = []ReadinessCheck{
		{
			Name:  "database",
			Mode:  ReadinessModeRequired,
			Probe: func(_ context.Context) error { return errors.New("dial failed") },
		},
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/ready", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusServiceUnavailable)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "SERVICE_UNAVAILABLE")
}

// TestNewServer_APIPrefixBaseline_NotFoundEnvelope ensures /api/v1 baseline exists and returns standard not found envelope.
func TestNewServer_APIPrefixBaseline_NotFoundEnvelope(t *testing.T) {
	srv := buildTestServer(t)
	userFixture := fixtures.SyntheticUser()

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/api/v1/ping", nil, map[string]string{
		"x-request-id": "req-api-v1-baseline",
		"Origin":       "http://localhost:5173",
		"X-Test-User":  userFixture.ID,
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusNotFound)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "NOT_FOUND")
}

func TestNewServer_MetricsRoute_ExposedWhenObservabilityEnabled(t *testing.T) {
	cfg := testServerConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := NewServer(cfg, logger, WithObservability(observability.NewRecorder()))
	if err != nil {
		t.Fatalf("failed to build server with observability: %v", err)
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/live", nil, map[string]string{
		"Origin": "http://localhost:5173",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)

	metricsResp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/metrics", nil, nil)
	httpharness.RequireStatus(t, metricsResp.StatusCode, fiber.StatusOK)
	if !strings.Contains(string(metricsResp.Body), "recova_http_requests_total") {
		t.Fatalf("expected recova metrics in output, got: %s", string(metricsResp.Body))
	}
}

func TestAuthRequestLimiter_RejectsAfterLimit(t *testing.T) {
	app := newLimiterTestApp()
	cfg := testServerConfig()
	cfg.Security.RateLimit.WindowMs = 60000
	cfg.Security.RateLimit.AuthMax = 1

	app.Use(authRequestLimiter(cfg))
	app.Post("/api/v1/auth/google", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(response.Success("ok", fiber.Map{"ok": true}, nil))
	})

	first := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/google", map[string]any{"token": "x"}, nil)
	httpharness.RequireStatus(t, first.StatusCode, fiber.StatusOK)

	second := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/google", map[string]any{"token": "x"}, nil)
	httpharness.RequireStatus(t, second.StatusCode, fiber.StatusTooManyRequests)
	httpharness.RequireErrorEnvelope(t, second.JSON, "RATE_LIMITED")
}

func buildTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := testServerConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := NewServer(cfg, logger)
	if err != nil {
		t.Fatalf("failed to build server: %v", err)
	}

	return srv
}

func testServerConfig() config.Config {
	return config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-test",
			AppEnv:    "test",
			Port:      "3000",
			APIPrefix: "/api/v1",
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
}

func newLimiterTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
