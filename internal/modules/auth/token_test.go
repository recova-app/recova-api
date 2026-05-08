package auth

import (
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

func TestTokenManager_IssueAndParseAccessToken(t *testing.T) {
	manager := newTestTokenManager()

	raw, session, err := manager.IssueAccessToken("user-1")
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	if raw == "" {
		t.Fatal("expected raw access token")
	}
	if session.AccessToken == "" || session.TokenType != "Bearer" {
		t.Fatalf("unexpected session payload: %#v", session)
	}

	claims, err := manager.ParseAccessToken(raw)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("unexpected subject: %s", claims.Subject)
	}
}

func TestTokenManager_ParseAccessToken_RejectsRefreshToken(t *testing.T) {
	manager := newTestTokenManager()

	raw, _, err := manager.IssueRefreshToken("user-1")
	if err != nil {
		t.Fatalf("issue refresh token: %v", err)
	}

	if _, err := manager.ParseAccessToken(raw); err == nil {
		t.Fatal("expected parse access token to fail for refresh token")
	}
}

func TestTokenManager_RefreshCookie_SetAndExpire(t *testing.T) {
	manager := newTestTokenManager()

	cookie := manager.RefreshCookie("refresh-token")
	if cookie.Name != "recova_refresh" {
		t.Fatalf("unexpected cookie name: %s", cookie.Name)
	}
	if cookie.Value != "refresh-token" {
		t.Fatalf("unexpected cookie value: %s", cookie.Value)
	}

	expired := manager.ExpiredRefreshCookie()
	if expired.MaxAge >= 0 {
		t.Fatalf("expected expired cookie max-age negative, got: %d", expired.MaxAge)
	}
}

func newTestTokenManager() *TokenManager {
	cfg := config.Config{
		Application: config.ApplicationConfig{
			AppName: "recova-auth-test",
		},
		Auth: config.AuthConfig{
			JWTSecret:     "test-jwt-secret-1234567890",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 24 * time.Hour,
			GoogleClient:  "google-client-test",
			Cookie: config.CookieConfig{
				Name:     "recova_refresh",
				Secure:   false,
				SameSite: "lax",
			},
		},
	}
	return NewTokenManager(cfg)
}
