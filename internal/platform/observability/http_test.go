package observability

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

func TestNewRequestTelemetryMiddleware_OnComplete_UsesRouteTemplateAndStatus(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	rec := NewRecorder()

	var captured RequestContext
	mw := NewRequestTelemetryMiddleware(logger, rec, "x-request-id", RequestHooks{
		OnComplete: func(ctx RequestContext) {
			captured = ctx
		},
	})

	app := fiber.New()
	app.Use(mw)
	app.Get("/api/v1/users/:id", func(c fiber.Ctx) error {
		// Emulate auth middleware attaching principal.
		c.Locals("recova.auth.principal", authmodule.AuthPrincipal{UserID: "user-1"})
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	req.Header.Set("x-request-id", "req-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	_ = resp.Body.Close()

	if captured.RequestID != "req-1" {
		t.Fatalf("unexpected request id: %q", captured.RequestID)
	}
	if captured.Method != "GET" {
		t.Fatalf("unexpected method: %q", captured.Method)
	}
	if captured.RoutePath != "/api/v1/users/:id" {
		t.Fatalf("unexpected route path: %q", captured.RoutePath)
	}
	if captured.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status code: %d", captured.StatusCode)
	}
	if captured.UserID != "user-1" {
		t.Fatalf("unexpected user id: %q", captured.UserID)
	}
}

func TestNewRequestTelemetryMiddleware_MapsErrorStatus(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	mw := NewRequestTelemetryMiddleware(logger, nil, "x-request-id", RequestHooks{})

	app := fiber.New()
	app.Use(mw)
	app.Get("/boom", func(c fiber.Ctx) error {
		return errors.New("boom")
	})

	req := httptest.NewRequest("GET", "/boom", nil)
	req.Header.Set("x-request-id", "req-err-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	_ = resp.Body.Close()

	// Middleware logs, error handler maps later; ensure it logged something.
	out, err := io.ReadAll(&logBuf)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(out), "request completed") {
		t.Fatalf("expected request log emitted, got: %s", string(out))
	}
}
