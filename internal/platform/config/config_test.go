// Package config tests configuration loading and validation behavior.
package config

import (
	"strings"
	"testing"
	"time"
)

// TestLoad_ValidEnv_ReturnsConfig ensures strict valid env can be loaded.
func TestLoad_ValidEnv_ReturnsConfig(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Application.AppName != "recova-backend-v2" {
		t.Fatalf("unexpected app name: %s", cfg.Application.AppName)
	}

	if cfg.Application.AppEnv != "local" {
		t.Fatalf("unexpected app env: %s", cfg.Application.AppEnv)
	}

	if cfg.Auth.JWTAccessTTL != 15*time.Minute {
		t.Fatalf("unexpected access ttl: %s", cfg.Auth.JWTAccessTTL)
	}

	if cfg.Logger.Level != "info" {
		t.Fatalf("unexpected logger level: %s", cfg.Logger.Level)
	}
}

// TestLoad_EmptyRequiredEnv_ReturnsError ensures required env cannot be empty.
func TestLoad_EmptyRequiredEnv_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_NAME", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "APP_NAME") {
		t.Fatalf("expected APP_NAME error, got: %v", err)
	}
}

// TestLoad_InvalidEnum_ReturnsError ensures enum validation fails fast.
func TestLoad_InvalidEnum_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "dev")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV issue, got: %v", err)
	}
}

// TestLoad_InvalidNumericValue_ReturnsError ensures numeric/range validation works.
func TestLoad_InvalidNumericValue_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("AI_TIMEOUT_MS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "AI_TIMEOUT_MS") {
		t.Fatalf("expected AI_TIMEOUT_MS issue, got: %v", err)
	}
}

// TestLoad_InvalidDuration_ReturnsError ensures duration format is enforced.
func TestLoad_InvalidDuration_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("JWT_ACCESS_TTL", "15minutes")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "JWT_ACCESS_TTL") {
		t.Fatalf("expected JWT_ACCESS_TTL issue, got: %v", err)
	}
}

// TestLoad_InvalidURL_ReturnsError ensures URL validation is enforced.
func TestLoad_InvalidURL_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_URL", "localhost:3000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "APP_URL") {
		t.Fatalf("expected APP_URL issue, got: %v", err)
	}
}

// TestLoad_FallbackPairMismatch_ReturnsError ensures fallback provider/model pair is consistent.
func TestLoad_FallbackPairMismatch_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("AI_FALLBACK_PROVIDER", "openai-compatible")
	t.Setenv("AI_FALLBACK_MODEL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "AI_FALLBACK_MODEL") {
		t.Fatalf("expected fallback model issue, got: %v", err)
	}
}

// TestConfig_RedactedSummary_DoesNotExposeSecret ensures startup summary never leaks secrets.
func TestConfig_RedactedSummary_DoesNotExposeSecret(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	summary := cfg.RedactedSummary()
	text := summary["database"].(map[string]any)["url"].(string)
	if text != redactedSecret {
		t.Fatalf("expected redacted database url, got: %s", text)
	}

	authSummary := summary["auth"].(map[string]any)
	if authSummary["jwtSecret"].(string) != redactedSecret {
		t.Fatalf("expected jwt secret redacted, got: %v", authSummary["jwtSecret"])
	}
}

func setValidEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_NAME", "recova-backend-v2")
	t.Setenv("APP_ENV", "local")
	t.Setenv("NODE_ENV", "development")
	t.Setenv("PORT", "3000")
	t.Setenv("API_PREFIX", "/api/v1")
	t.Setenv("APP_URL", "http://localhost:3000")
	t.Setenv("DOCS_URL", "https://docs.recova.app")

	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/recova_db?sslmode=disable")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/recova_db?sslmode=disable")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "25")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "10")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME_SEC", "300")
	t.Setenv("DATABASE_SSL_MODE", "disable")

	t.Setenv("JWT_SECRET", "replace-with-strong-secret")
	t.Setenv("JWT_ACCESS_TTL", "15m")
	t.Setenv("JWT_REFRESH_TTL", "7d")
	t.Setenv("GOOGLE_CLIENT_ID", "replace-with-google-client-id")
	t.Setenv("AUTH_COOKIE_NAME", "recova_refresh")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("AUTH_COOKIE_SAME_SITE", "lax")
	t.Setenv("AUTH_COOKIE_DOMAIN", "")

	t.Setenv("AI_PROVIDER", "gemini")
	t.Setenv("AI_MODEL", "gemini-2.0-flash")
	t.Setenv("AI_API_KEY", "replace-with-provider-key")
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_TIMEOUT_MS", "10000")
	t.Setenv("AI_FALLBACK_PROVIDER", "")
	t.Setenv("AI_FALLBACK_MODEL", "")

	t.Setenv("CORS_ORIGINS", "http://localhost:5173,http://localhost:4173")
	t.Setenv("RATE_LIMIT_WINDOW_MS", "60000")
	t.Setenv("RATE_LIMIT_MAX", "120")
	t.Setenv("AUTH_RATE_LIMIT_MAX", "10")
	t.Setenv("AI_RATE_LIMIT_MAX", "20")
	t.Setenv("REQUEST_BODY_LIMIT", "1mb")

	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REQUEST_ID_HEADER", "x-request-id")
	t.Setenv("HEALTH_CHECK_TIMEOUT_MS", "2000")
}
