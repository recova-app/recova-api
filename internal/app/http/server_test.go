// Package http tests baseline HTTP runtime behavior and middleware contract.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
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
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Origin", "http://localhost:5173")

			resp, payload := doRequest(t, srv, req)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: %d", resp.StatusCode)
			}

			success, ok := payload["success"].(bool)
			if !ok || !success {
				t.Fatalf("expected success envelope, got: %v", payload)
			}
		})
	}
}

// TestNewServer_NotFound_UsesStandardErrorEnvelope ensures unknown routes return API-standard error payload.
func TestNewServer_NotFound_UsesStandardErrorEnvelope(t *testing.T) {
	srv := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	req.Header.Set("x-request-id", "req-notfound-1")

	resp, payload := doRequest(t, srv, req)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got: %d", resp.StatusCode)
	}

	if payload["success"] != false {
		t.Fatalf("expected success=false, got: %v", payload["success"])
	}

	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND code, got: %v", errorPayload["code"])
	}

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

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	resp, payload := doRequest(t, srv, req)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got: %d", resp.StatusCode)
	}

	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR code, got: %v", errorPayload["code"])
	}
}

// TestNewServer_RequestIDAndSecurityHeaders ensures request id and security/CORS headers are present.
func TestNewServer_RequestIDAndSecurityHeaders(t *testing.T) {
	srv := buildTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("x-request-id", "req-header-1")

	resp, _ := doRequest(t, srv, req)

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

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	resp, payload := doRequest(t, srv, req)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got: %d", resp.StatusCode)
	}

	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE code, got: %v", errorPayload["code"])
	}
}

func buildTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-test",
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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := NewServer(cfg, logger)
	if err != nil {
		t.Fatalf("failed to build server: %v", err)
	}

	return srv
}

func doRequest(t *testing.T, srv *Server, req *http.Request) (*http.Response, map[string]any) {
	t.Helper()

	resp, err := srv.app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed reading response body: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		t.Fatalf("failed parsing response payload: %v", err)
	}

	return resp, payload
}
