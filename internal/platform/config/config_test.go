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

// TestLoad_ProductionDatabaseSSLModeDisabled_ReturnsError ensures production rejects disabled SSL.
func TestLoad_ProductionDatabaseSSLModeDisabled_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "disable")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=disable")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=disable")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for disabled ssl in production")
	}
	if !strings.Contains(err.Error(), "DATABASE_SSL_MODE") {
		t.Fatalf("expected DATABASE_SSL_MODE issue, got: %v", err)
	}
}

// TestLoad_ProductionDatabaseURLWithoutSSLMode_ReturnsError ensures production requires explicit sslmode parameter.
func TestLoad_ProductionDatabaseURLWithoutSSLMode_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for missing sslmode in database url")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL issue, got: %v", err)
	}
}

// TestLoad_DatabaseURLSSLModeMismatch_ReturnsError ensures url sslmode param matches setting.
func TestLoad_DatabaseURLSSLModeMismatch_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/recova_db?sslmode=disable")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/recova_db?sslmode=require")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for sslmode mismatch")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL issue, got: %v", err)
	}
}

// TestLoad_DirectDatabaseURLSSLModeMismatch_ReturnsError ensures direct url sslmode param matches setting.
func TestLoad_DirectDatabaseURLSSLModeMismatch_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DATABASE_SSL_MODE", "disable")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/recova_db?sslmode=disable")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/recova_db?sslmode=require")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for direct database url sslmode mismatch")
	}
	if !strings.Contains(err.Error(), "DIRECT_DATABASE_URL") {
		t.Fatalf("expected DIRECT_DATABASE_URL issue, got: %v", err)
	}
}

// TestLoad_ProductionDatabaseSSLModeValid_ReturnsConfig ensures production accepts valid TLS config.
func TestLoad_ProductionDatabaseSSLModeValid_ReturnsConfig(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
	if cfg.Database.SSLMode != "require" {
		t.Fatalf("unexpected ssl mode: %s", cfg.Database.SSLMode)
	}
}

// TestLoad_ProductionAIBaseURLNonHTTPS_ReturnsError ensures production rejects non-https AI endpoints.
func TestLoad_ProductionAIBaseURLNonHTTPS_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("AI_PROVIDER", "gemini")
	t.Setenv("AI_BASE_URL", "http://generativelanguage.googleapis.com/v1beta")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for non-https ai endpoint in production")
	}
	if !strings.Contains(err.Error(), "AI_BASE_URL") || !strings.Contains(err.Error(), "https") {
		t.Fatalf("expected AI_BASE_URL https error, got: %v", err)
	}
}

// TestLoad_ProductionAIBaseURLUntrustedHost_ReturnsError ensures production rejects untrusted AI host.
func TestLoad_ProductionAIBaseURLUntrustedHost_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("AI_PROVIDER", "openai-compatible")
	t.Setenv("AI_BASE_URL", "https://untrusted-host.example.com/v1")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for untrusted ai host in production")
	}
	if !strings.Contains(err.Error(), "AI_BASE_URL") || !strings.Contains(err.Error(), "allowed host") {
		t.Fatalf("expected AI_BASE_URL allowed host error, got: %v", err)
	}
}

// TestLoad_ProductionAIBaseURLTrustedHost_ReturnsConfig ensures production accepts trusted AI hosts.
func TestLoad_ProductionAIBaseURLTrustedHost_ReturnsConfig(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("AI_PROVIDER", "openai-compatible")
	t.Setenv("AI_BASE_URL", "https://api.openai.com/v1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
	if cfg.AI.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected ai base url: %s", cfg.AI.BaseURL)
	}
}

// TestLoad_StagingAIBaseURLNonHTTPS_ReturnsError ensures staging rejects non-https AI endpoints.
func TestLoad_StagingAIBaseURLNonHTTPS_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "staging")
	t.Setenv("AI_PROVIDER", "gemini")
	t.Setenv("AI_BASE_URL", "http://generativelanguage.googleapis.com/v1beta")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for non-https ai endpoint in staging")
	}
	if !strings.Contains(err.Error(), "AI_BASE_URL") || !strings.Contains(err.Error(), "https") {
		t.Fatalf("expected AI_BASE_URL https error, got: %v", err)
	}
}

// TestLoad_StagingAIBaseURLUntrustedHost_ReturnsError ensures staging rejects untrusted AI host.
func TestLoad_StagingAIBaseURLUntrustedHost_ReturnsError(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "staging")
	t.Setenv("AI_PROVIDER", "gemini")
	t.Setenv("AI_BASE_URL", "https://untrusted-proxy.example.com/v1beta")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for untrusted ai host in staging")
	}
	if !strings.Contains(err.Error(), "AI_BASE_URL") || !strings.Contains(err.Error(), "allowed host") {
		t.Fatalf("expected AI_BASE_URL allowed host error, got: %v", err)
	}
}

// TestLoad_HeterogeneousFallbackWithCustomBaseURL_AllowedForPrimary ensures primary base URL does not block heterogeneous fallback.
func TestLoad_HeterogeneousFallbackWithCustomBaseURL_AllowedForPrimary(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_SSL_MODE", "require")
	t.Setenv("DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("DIRECT_DATABASE_URL", "postgresql://postgres:postgres@db.production:5432/recova_db?sslmode=require")
	t.Setenv("AI_PROVIDER", "gemini")
	t.Setenv("AI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta")
	t.Setenv("AI_FALLBACK_PROVIDER", "openai-compatible")
	t.Setenv("AI_FALLBACK_MODEL", "gpt-4.1-mini")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected valid config for heterogeneous fallback, got: %v", err)
	}
	if cfg.AI.BaseURL != "https://generativelanguage.googleapis.com/v1beta" {
		t.Fatalf("unexpected ai base url: %s", cfg.AI.BaseURL)
	}
}

// TestLoad_LocalAIBaseURLCustomHost_ReturnsConfig ensures local environment allows custom mock hosts.
func TestLoad_LocalAIBaseURLCustomHost_ReturnsConfig(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "local")
	t.Setenv("AI_PROVIDER", "openai-compatible")
	t.Setenv("AI_BASE_URL", "http://localhost:8080/v1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected valid config in local, got: %v", err)
	}
	if cfg.AI.BaseURL != "http://localhost:8080/v1" {
		t.Fatalf("unexpected ai base url: %s", cfg.AI.BaseURL)
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
