// Package http tests baseline HTTP server routes used for local verification.
package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

// TestNewServer_BaseHealthRoutes_ReturnOK ensures baseline health routes are registered.
func TestNewServer_BaseHealthRoutes_ReturnOK(t *testing.T) {
	cfg := config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-test",
			AppEnv:    "test",
			Port:      "3000",
			APIPrefix: "/api/v1",
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(cfg, logger)

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
			resp, err := srv.app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: %d", resp.StatusCode)
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed reading response body: %v", err)
			}

			var payload map[string]any
			if err := json.Unmarshal(bodyBytes, &payload); err != nil {
				t.Fatalf("failed parsing response payload: %v", err)
			}

			success, ok := payload["success"].(bool)
			if !ok || !success {
				t.Fatalf("expected success envelope, got: %v", payload)
			}
		})
	}
}
