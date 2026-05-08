// Package config provides runtime configuration loading and validation.
package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const redactedSecret = "[REDACTED_SECRET]"

var (
	allowedAppEnv       = []string{"local", "test", "staging", "production"}
	allowedNodeEnv      = []string{"development", "test", "production"}
	allowedDatabaseSSL  = []string{"disable", "require", "verify-ca", "verify-full"}
	allowedCookieSame   = []string{"lax", "strict", "none"}
	allowedAIProvider   = []string{"gemini", "openai-compatible"}
	allowedLogLevel     = []string{"debug", "info", "warn", "error"}
	allowedFallbackMode = []string{"gemini", "openai-compatible"}
)

// Config holds full runtime configuration grouped by concern.
type Config struct {
	Application   ApplicationConfig
	Database      DatabaseConfig
	Auth          AuthConfig
	AI            AIConfig
	Security      SecurityConfig
	Observability ObservabilityConfig
	Logger        LoggerConfig
}

// ApplicationConfig stores application runtime settings.
type ApplicationConfig struct {
	AppName   string
	AppEnv    string
	NodeEnv   string
	Port      string
	APIPrefix string
	AppURL    string
	DocsURL   string
}

// DatabaseConfig stores PostgreSQL connectivity and pooling settings.
type DatabaseConfig struct {
	URL                string
	DirectURL          string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetimeSec int
	SSLMode            string
}

// AuthConfig stores authentication settings.
type AuthConfig struct {
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
	GoogleClient  string
	Cookie        CookieConfig
}

// CookieConfig stores refresh cookie policy.
type CookieConfig struct {
	Name     string
	Secure   bool
	SameSite string
	Domain   string
}

// AIConfig stores AI provider settings.
type AIConfig struct {
	Provider         string
	Model            string
	APIKey           string
	BaseURL          string
	TimeoutMs        int
	FallbackProvider string
	FallbackModel    string
}

// SecurityConfig stores security and traffic guard settings.
type SecurityConfig struct {
	CORSOrigins      []string
	RequestBodyLimit string
	RateLimit        RateLimitConfig
}

// RateLimitConfig stores global and endpoint-specific rate-limit rules.
type RateLimitConfig struct {
	WindowMs int
	Max      int
	AuthMax  int
	AIMax    int
}

// ObservabilityConfig stores health and tracing header settings.
type ObservabilityConfig struct {
	RequestIDHeader      string
	HealthCheckTimeoutMs int
}

// LoggerConfig stores structured logger settings.
type LoggerConfig struct {
	Level     string
	SlogLevel slog.Level
}

// Load reads and validates runtime configuration from environment variables.
func Load() (Config, error) {
	v := &validator{}
	cfg := Config{
		Application: ApplicationConfig{
			AppName:   v.requiredString("APP_NAME"),
			AppEnv:    v.requiredEnum("APP_ENV", allowedAppEnv),
			NodeEnv:   v.requiredEnum("NODE_ENV", allowedNodeEnv),
			Port:      v.requiredString("PORT"),
			APIPrefix: v.requiredString("API_PREFIX"),
			AppURL:    v.requiredAbsoluteURL("APP_URL"),
			DocsURL:   v.requiredAbsoluteURL("DOCS_URL"),
		},
		Database: DatabaseConfig{
			URL:                v.requiredDatabaseURL("DATABASE_URL"),
			DirectURL:          v.optionalDatabaseURL("DIRECT_DATABASE_URL"),
			MaxOpenConns:       v.requiredInt("DATABASE_MAX_OPEN_CONNS", 1, 1000),
			MaxIdleConns:       v.requiredInt("DATABASE_MAX_IDLE_CONNS", 1, 1000),
			ConnMaxLifetimeSec: v.requiredInt("DATABASE_CONN_MAX_LIFETIME_SEC", 1, 86400),
			SSLMode:            v.requiredEnum("DATABASE_SSL_MODE", allowedDatabaseSSL),
		},
		Auth: AuthConfig{
			JWTSecret:     v.requiredSecret("JWT_SECRET", 16),
			JWTAccessTTL:  v.requiredDuration("JWT_ACCESS_TTL"),
			JWTRefreshTTL: v.requiredDuration("JWT_REFRESH_TTL"),
			GoogleClient:  v.requiredString("GOOGLE_CLIENT_ID"),
			Cookie: CookieConfig{
				Name:     v.requiredString("AUTH_COOKIE_NAME"),
				Secure:   v.requiredBool("AUTH_COOKIE_SECURE"),
				SameSite: v.requiredEnum("AUTH_COOKIE_SAME_SITE", allowedCookieSame),
				Domain:   v.optionalString("AUTH_COOKIE_DOMAIN"),
			},
		},
		AI: AIConfig{
			Provider:         v.requiredEnum("AI_PROVIDER", allowedAIProvider),
			Model:            v.requiredString("AI_MODEL"),
			APIKey:           v.requiredSecret("AI_API_KEY", 8),
			BaseURL:          v.optionalAbsoluteURL("AI_BASE_URL"),
			TimeoutMs:        v.requiredInt("AI_TIMEOUT_MS", 1, 60000),
			FallbackProvider: v.optionalEnum("AI_FALLBACK_PROVIDER", allowedFallbackMode),
			FallbackModel:    v.optionalString("AI_FALLBACK_MODEL"),
		},
		Security: SecurityConfig{
			CORSOrigins:      v.requiredCSVAbsoluteURLs("CORS_ORIGINS"),
			RequestBodyLimit: v.requiredString("REQUEST_BODY_LIMIT"),
			RateLimit: RateLimitConfig{
				WindowMs: v.requiredInt("RATE_LIMIT_WINDOW_MS", 1, 3600000),
				Max:      v.requiredInt("RATE_LIMIT_MAX", 1, 100000),
				AuthMax:  v.requiredInt("AUTH_RATE_LIMIT_MAX", 1, 100000),
				AIMax:    v.requiredInt("AI_RATE_LIMIT_MAX", 1, 100000),
			},
		},
		Observability: ObservabilityConfig{
			RequestIDHeader:      v.requiredString("REQUEST_ID_HEADER"),
			HealthCheckTimeoutMs: v.requiredInt("HEALTH_CHECK_TIMEOUT_MS", 1, 60000),
		},
		Logger: LoggerConfig{
			Level: v.requiredEnum("LOG_LEVEL", allowedLogLevel),
		},
	}

	cfg.Logger.SlogLevel = parseSlogLevel(cfg.Logger.Level)

	if !strings.HasPrefix(cfg.Application.APIPrefix, "/") {
		v.addIssue("API_PREFIX", "must start with '/'")
	}

	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		v.addIssue("DATABASE_MAX_IDLE_CONNS", "must not be greater than DATABASE_MAX_OPEN_CONNS")
	}

	if cfg.AI.FallbackProvider == "" && cfg.AI.FallbackModel != "" {
		v.addIssue("AI_FALLBACK_PROVIDER", "is required when AI_FALLBACK_MODEL is set")
	}

	if cfg.AI.FallbackProvider != "" && cfg.AI.FallbackModel == "" {
		v.addIssue("AI_FALLBACK_MODEL", "is required when AI_FALLBACK_PROVIDER is set")
	}

	if v.hasIssue() {
		return Config{}, v.toError()
	}

	return cfg, nil
}

// RedactedSummary returns a safe configuration snapshot for logs.
func (c Config) RedactedSummary() map[string]any {
	return map[string]any{
		"application": map[string]any{
			"appName":   c.Application.AppName,
			"appEnv":    c.Application.AppEnv,
			"nodeEnv":   c.Application.NodeEnv,
			"port":      c.Application.Port,
			"apiPrefix": c.Application.APIPrefix,
			"appURL":    c.Application.AppURL,
			"docsURL":   c.Application.DocsURL,
		},
		"database": map[string]any{
			"url":                 redactedSecret,
			"directUrlConfigured": c.Database.DirectURL != "",
			"maxOpenConns":        c.Database.MaxOpenConns,
			"maxIdleConns":        c.Database.MaxIdleConns,
			"connMaxLifetimeSec":  c.Database.ConnMaxLifetimeSec,
			"sslMode":             c.Database.SSLMode,
		},
		"auth": map[string]any{
			"jwtSecret":       redactedSecret,
			"jwtAccessTTL":    c.Auth.JWTAccessTTL.String(),
			"jwtRefreshTTL":   c.Auth.JWTRefreshTTL.String(),
			"googleClientID":  c.Auth.GoogleClient,
			"cookieName":      c.Auth.Cookie.Name,
			"cookieSecure":    c.Auth.Cookie.Secure,
			"cookieSameSite":  c.Auth.Cookie.SameSite,
			"cookieHasDomain": c.Auth.Cookie.Domain != "",
		},
		"ai": map[string]any{
			"provider":         c.AI.Provider,
			"model":            c.AI.Model,
			"apiKey":           redactedSecret,
			"hasBaseURL":       c.AI.BaseURL != "",
			"timeoutMs":        c.AI.TimeoutMs,
			"fallbackProvider": c.AI.FallbackProvider,
			"fallbackModel":    c.AI.FallbackModel,
		},
		"security": map[string]any{
			"corsOrigins":      c.Security.CORSOrigins,
			"requestBodyLimit": c.Security.RequestBodyLimit,
			"rateLimitWindow":  c.Security.RateLimit.WindowMs,
			"rateLimitMax":     c.Security.RateLimit.Max,
			"authRateLimitMax": c.Security.RateLimit.AuthMax,
			"aiRateLimitMax":   c.Security.RateLimit.AIMax,
		},
		"observability": map[string]any{
			"logLevel":             c.Logger.Level,
			"requestIDHeader":      c.Observability.RequestIDHeader,
			"healthCheckTimeoutMs": c.Observability.HealthCheckTimeoutMs,
		},
	}
}

type validator struct {
	issues []string
}

func (v *validator) requiredString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		v.addIssue(key, "is required")
		return ""
	}

	value = strings.TrimSpace(value)
	if value == "" {
		v.addIssue(key, "is required and cannot be empty")
	}

	return value
}

func (v *validator) optionalString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return ""
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	return value
}

func (v *validator) requiredSecret(key string, minLen int) string {
	value := v.requiredString(key)
	if value == "" {
		return ""
	}

	if len(value) < minLen {
		v.addIssue(key, fmt.Sprintf("minimum %d characters", minLen))
	}

	return value
}

func (v *validator) requiredEnum(key string, allowed []string) string {
	value := strings.ToLower(v.requiredString(key))
	if value == "" {
		return ""
	}

	if !slices.Contains(allowed, value) {
		v.addIssue(key, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
	}

	return value
}

func (v *validator) optionalEnum(key string, allowed []string) string {
	value := strings.ToLower(v.optionalString(key))
	if value == "" {
		return ""
	}

	if !slices.Contains(allowed, value) {
		v.addIssue(key, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
	}

	return value
}

func (v *validator) requiredBool(key string) bool {
	value := v.requiredString(key)
	if value == "" {
		return false
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		v.addIssue(key, "must be a valid boolean")
		return false
	}

	return parsed
}

func (v *validator) requiredInt(key string, min int, max int) int {
	value := v.requiredString(key)
	if value == "" {
		return 0
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		v.addIssue(key, "must be a valid integer")
		return 0
	}

	if parsed < min || parsed > max {
		v.addIssue(key, fmt.Sprintf("must be in range %d..%d", min, max))
	}

	return parsed
}

func (v *validator) requiredDuration(key string) time.Duration {
	value := v.requiredString(key)
	if value == "" {
		return 0
	}

	parsed, err := parseDuration(value)
	if err != nil {
		v.addIssue(key, "invalid duration format")
		return 0
	}

	if parsed <= 0 {
		v.addIssue(key, "duration must be greater than 0")
	}

	return parsed
}

func parseDuration(raw string) (time.Duration, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if strings.HasSuffix(value, "d") {
		daysRaw := strings.TrimSuffix(value, "d")
		days, err := strconv.Atoi(daysRaw)
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("invalid day duration")
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(value)
}

func (v *validator) requiredAbsoluteURL(key string) string {
	value := v.requiredString(key)
	if value == "" {
		return ""
	}

	if !isAbsoluteURL(value) {
		v.addIssue(key, "must be a valid absolute URL")
	}

	return value
}

func (v *validator) optionalAbsoluteURL(key string) string {
	value := v.optionalString(key)
	if value == "" {
		return ""
	}

	if !isAbsoluteURL(value) {
		v.addIssue(key, "must be a valid absolute URL")
	}

	return value
}

func (v *validator) requiredDatabaseURL(key string) string {
	value := v.requiredString(key)
	if value == "" {
		return ""
	}

	if !isDatabaseURL(value) {
		v.addIssue(key, "must be a valid database URL")
	}

	return value
}

func (v *validator) optionalDatabaseURL(key string) string {
	value := v.optionalString(key)
	if value == "" {
		return ""
	}

	if !isDatabaseURL(value) {
		v.addIssue(key, "must be a valid database URL")
	}

	return value
}

func (v *validator) requiredCSVAbsoluteURLs(key string) []string {
	value := v.requiredString(key)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}

		if !isAbsoluteURL(origin) {
			v.addIssue(key, fmt.Sprintf("invalid origin: %s", origin))
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		v.addIssue(key, "at least one valid origin is required")
	}

	return origins
}

func (v *validator) addIssue(key string, message string) {
	v.issues = append(v.issues, fmt.Sprintf("%s: %s", key, message))
}

func (v *validator) hasIssue() bool {
	return len(v.issues) > 0
}

func (v *validator) toError() error {
	return fmt.Errorf("CONFIG_VALIDATION_ERROR: %s", strings.Join(v.issues, "; "))
}

func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func isAbsoluteURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}

	return parsed.Scheme != "" && parsed.Host != ""
}

func isDatabaseURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}

	return parsed.Scheme != "" && parsed.Host != ""
}
