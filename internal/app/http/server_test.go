// Package http tests baseline HTTP runtime behavior and middleware contract.
package http

import (
	"bytes"
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

func TestNewServer_OpenAPIArtifactRoute_ReturnsGeneratedSpec(t *testing.T) {
	srv := buildTestServer(t)

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/openapi.yaml", nil, map[string]string{
		"Origin": "http://localhost:5173",
	})

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/yaml") {
		t.Fatalf("expected openapi content-type application/yaml, got: %q", contentType)
	}

	body := string(resp.Body)
	if !strings.Contains(body, "openapi: 3.1.0") {
		t.Fatalf("expected openapi yaml body, got: %s", httpharness.DebugBody(resp))
	}
}

func TestNewServer_ScalarDocsRoute_ReturnsHTMLWithCSP(t *testing.T) {
	srv := buildTestServer(t)

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/docs/api", nil, map[string]string{
		"Origin": "http://localhost:5173",
	})

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("expected text/html content type, got: %q", contentType)
	}

	csp := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://cdn.jsdelivr.net/npm/@scalar/api-reference") {
		t.Fatalf("expected scalar script source on csp header, got: %q", csp)
	}
	if !strings.Contains(csp, "'nonce-") {
		t.Fatalf("expected nonce in csp header, got: %q", csp)
	}

	body := string(resp.Body)
	if !strings.Contains(body, "@scalar/api-reference") {
		t.Fatalf("expected scalar script tag in html body, got: %s", httpharness.DebugBody(resp))
	}
	if !strings.Contains(body, "url: '/openapi.yaml'") {
		t.Fatalf("expected openapi url on scalar html initializer, got: %s", httpharness.DebugBody(resp))
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
	if errorPayload["request_id"] != "req-notfound-1" {
		t.Fatalf("expected request id propagation, got: %v", errorPayload["request_id"])
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

// TestNewServer_ReadinessFailure_ReturnsServiceUnavailable ensures failed required dependency returns 503 without leaking internal errors.
func TestNewServer_ReadinessFailure_ReturnsServiceUnavailable(t *testing.T) {
	srv := buildTestServer(t)
	var logBuf bytes.Buffer
	srv.logger = slog.New(slog.NewJSONHandler(&logBuf, nil))

	rawError := "dial tcp 127.0.0.1:5432: connection refused"
	srv.readinessChecks = []ReadinessCheck{
		{
			Name:  "database",
			Mode:  ReadinessModeRequired,
			Probe: func(_ context.Context) error { return errors.New(rawError) },
		},
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/ready", nil, map[string]string{
		"x-request-id": "req-readiness-fail",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusServiceUnavailable)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "SERVICE_UNAVAILABLE")

	if strings.Contains(string(resp.Body), rawError) {
		t.Fatalf("readiness response leaked internal error detail: %s", string(resp.Body))
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "readiness probe failed") {
		t.Fatalf("expected structured log 'readiness probe failed', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "req-readiness-fail") {
		t.Fatalf("expected structured log to include request_id 'req-readiness-fail', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "database") {
		t.Fatalf("expected structured log to include dependency 'database', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, rawError) {
		t.Fatalf("expected structured log to include internal probe error detail, got: %s", logOutput)
	}

	errObj, ok := resp.JSON["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got: %#v", resp.JSON["error"])
	}
	details, ok := errObj["details"].(map[string]any)
	if !ok {
		t.Fatalf("expected details object, got: %#v", errObj["details"])
	}
	checks, ok := details["checks"].(map[string]any)
	if !ok {
		t.Fatalf("expected checks object, got: %#v", details["checks"])
	}
	dbCheck, ok := checks["database"].(map[string]any)
	if !ok {
		t.Fatalf("expected database check object, got: %#v", checks["database"])
	}
	if dbCheck["status"] != "down" {
		t.Fatalf("expected status down, got: %v", dbCheck["status"])
	}
	if dbCheck["message"] != "Dependency tidak sehat" {
		t.Fatalf("expected safe message, got: %v", dbCheck["message"])
	}
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
	rec := observability.NewRecorder()
	srv, err := NewServer(cfg, logger, WithObservability(rec))
	if err != nil {
		t.Fatalf("failed to build server with observability: %v", err)
	}

	rec.RecordAIFallback("gemini", nil)

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/live", nil, map[string]string{
		"Origin": "http://localhost:5173",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)

	metricsResp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/metrics", nil, nil)
	httpharness.RequireStatus(t, metricsResp.StatusCode, fiber.StatusOK)
	if !strings.Contains(string(metricsResp.Body), "recova_http_requests_total") {
		t.Fatalf("expected recova metrics in output, got: %s", string(metricsResp.Body))
	}
	if !strings.Contains(string(metricsResp.Body), "recova_ai_fallback_total") {
		t.Fatalf("expected recova_ai_fallback_total in output, got: %s", string(metricsResp.Body))
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

func TestNewServer_ProductionWithoutRequiredDatabaseProbe_ReturnsError(t *testing.T) {
	cfg := testServerConfig()
	cfg.Application.AppEnv = "production"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, err := NewServer(cfg, logger)
	if err == nil {
		t.Fatal("expected error when building production server without active database probe")
	}
	if !strings.Contains(err.Error(), "production environment") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewServer_ProductionWithRequiredDatabaseProbe_Succeeds(t *testing.T) {
	cfg := testServerConfig()
	cfg.Application.AppEnv = "production"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	srv, err := NewServer(cfg, logger, WithReadinessChecks([]ReadinessCheck{
		{
			Name:  "database",
			Mode:  ReadinessModeRequired,
			Probe: func(_ context.Context) error { return nil },
		},
	}))
	if err != nil {
		t.Fatalf("unexpected error building production server: %v", err)
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/ready", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)

	data := resp.JSON["data"].(map[string]any)
	checks := data["checks"].(map[string]any)
	dbCheck := checks["database"].(map[string]any)
	if dbCheck["message"] != "Koneksi database sehat" {
		t.Fatalf("expected healthy message 'Koneksi database sehat', got: %v", dbCheck["message"])
	}
}

func TestNewServer_ProductionDatabaseDown_ReturnsServiceUnavailable(t *testing.T) {
	cfg := testServerConfig()
	cfg.Application.AppEnv = "production"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	srv, err := NewServer(cfg, logger, WithReadinessChecks([]ReadinessCheck{
		{
			Name:    "database",
			Mode:    ReadinessModeRequired,
			Message: "Database offline",
			Probe:   func(_ context.Context) error { return errors.New("db connection down") },
		},
	}))
	if err != nil {
		t.Fatalf("unexpected error building production server: %v", err)
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/health/ready", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusServiceUnavailable)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "SERVICE_UNAVAILABLE")
}

func TestNewServer_WithRouteRegistrars(t *testing.T) {
	cfg := testServerConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	customRegistrar := RouteRegistrarFunc(func(router fiber.Router) {
		router.Get("/custom-endpoint", func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"custom": true})
		})
	})

	srv, err := NewServer(cfg, logger, WithRouteRegistrars(customRegistrar))
	if err != nil {
		t.Fatalf("failed to build server with route registrar: %v", err)
	}

	resp := httpharness.JSONRequest(t, srv.app, fiber.MethodGet, "/api/v1/custom-endpoint", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
}
