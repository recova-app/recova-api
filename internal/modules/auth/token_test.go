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
	if len(claims.Audience) == 0 || claims.Audience[0] != "http://localhost:3000" {
		t.Fatalf("unexpected audience: %#v", claims.Audience)
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

func TestTokenManager_RefreshCookie_AllAttributes(t *testing.T) {
	testCases := []struct {
		name     string
		secure   bool
		sameSite string
		domain   string
	}{
		{name: "dev lax without domain", secure: false, sameSite: "lax", domain: ""},
		{name: "prod strict with domain", secure: true, sameSite: "strict", domain: "recova.app"},
		{name: "none with domain", secure: true, sameSite: "none", domain: "api.recova.app"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				Application: config.ApplicationConfig{
					AppName: "recova-app",
					AppURL:  "https://recova.app",
				},
				Auth: config.AuthConfig{
					JWTSecret:     "secret-1234567890-test-key",
					JWTAccessTTL:  15 * time.Minute,
					JWTRefreshTTL: 7 * 24 * time.Hour,
					Cookie: config.CookieConfig{
						Name:     "recova_refresh",
						Secure:   tc.secure,
						SameSite: tc.sameSite,
						Domain:   tc.domain,
					},
				},
			}
			manager := NewTokenManager(cfg)
			start := time.Now()
			cookie := manager.RefreshCookie("sample-raw-token")

			if cookie.Name != "recova_refresh" {
				t.Fatalf("unexpected cookie name: %s", cookie.Name)
			}
			if cookie.Value != "sample-raw-token" {
				t.Fatalf("unexpected cookie value: %s", cookie.Value)
			}
			if cookie.Path != "/" {
				t.Fatalf("unexpected cookie path: %s", cookie.Path)
			}
			if !cookie.HTTPOnly {
				t.Fatal("expected httpOnly to be true")
			}
			if cookie.Secure != tc.secure {
				t.Fatalf("expected secure=%v, got=%v", tc.secure, cookie.Secure)
			}
			if cookie.SameSite != tc.sameSite {
				t.Fatalf("expected sameSite=%s, got=%s", tc.sameSite, cookie.SameSite)
			}
			if cookie.Domain != tc.domain {
				t.Fatalf("expected domain=%s, got=%s", tc.domain, cookie.Domain)
			}
			expectedMin := start.Add(7*24*time.Hour - 5*time.Second)
			expectedMax := time.Now().Add(7*24*time.Hour + 5*time.Second)
			if cookie.Expires.Before(expectedMin) || cookie.Expires.After(expectedMax) {
				t.Fatalf("cookie expires %v not in expected range [%v, %v]", cookie.Expires, expectedMin, expectedMax)
			}
		})
	}
}

func TestTokenManager_ExpiredRefreshCookie_AllAttributes(t *testing.T) {
	cfg := config.Config{
		Application: config.ApplicationConfig{AppName: "recova-app", AppURL: "https://recova.app"},
		Auth: config.AuthConfig{
			JWTSecret:     "secret-1234567890-test-key",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 24 * time.Hour,
			Cookie: config.CookieConfig{
				Name:     "recova_refresh",
				Secure:   true,
				SameSite: "strict",
				Domain:   "recova.app",
			},
		},
	}
	manager := NewTokenManager(cfg)
	cookie := manager.ExpiredRefreshCookie()

	if cookie.Name != "recova_refresh" {
		t.Fatalf("unexpected cookie name: %s", cookie.Name)
	}
	if cookie.Value != "" {
		t.Fatalf("expected empty value, got: %s", cookie.Value)
	}
	if cookie.Path != "/" {
		t.Fatalf("unexpected cookie path: %s", cookie.Path)
	}
	if !cookie.HTTPOnly {
		t.Fatal("expected httpOnly to be true")
	}
	if !cookie.Secure {
		t.Fatal("expected secure to be true")
	}
	if cookie.SameSite != "strict" {
		t.Fatalf("unexpected sameSite: %s", cookie.SameSite)
	}
	if cookie.Domain != "recova.app" {
		t.Fatalf("unexpected domain: %s", cookie.Domain)
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("expected max-age -1, got: %d", cookie.MaxAge)
	}
	if !cookie.Expires.Equal(time.Unix(1, 0)) {
		t.Fatalf("expected unix(1,0) expires, got: %v", cookie.Expires)
	}
}

func TestTokenManager_ParseRefreshToken_MalformedAndInvalid(t *testing.T) {
	manager := newTestTokenManager()

	cases := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "whitespace token", token: "   "},
		{name: "malformed text", token: "not-a-valid-jwt"},
		{name: "broken parts", token: "a.b.c"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := manager.ParseRefreshToken(tc.token); err == nil {
				t.Fatalf("expected parse refresh token error for %q", tc.name)
			}
		})
	}

	// Access token passed to ParseRefreshToken
	accessRaw, _, err := manager.IssueAccessToken("user-1")
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	if _, err := manager.ParseRefreshToken(accessRaw); err == nil {
		t.Fatal("expected parse refresh token to reject access token type")
	}

	// Token signed with different secret
	diffCfg := config.Config{
		Application: config.ApplicationConfig{AppName: "recova-auth-test", AppURL: "http://localhost:3000"},
		Auth: config.AuthConfig{
			JWTSecret:     "different-secret-0987654321",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 24 * time.Hour,
		},
	}
	diffManager := NewTokenManager(diffCfg)
	foreignToken, _, err := diffManager.IssueRefreshToken("user-1")
	if err != nil {
		t.Fatalf("issue foreign token: %v", err)
	}
	if _, err := manager.ParseRefreshToken(foreignToken); err == nil {
		t.Fatal("expected parse refresh token to reject foreign signed token")
	}

	// Expired refresh token
	expCfg := config.Config{
		Application: config.ApplicationConfig{AppName: "recova-auth-test", AppURL: "http://localhost:3000"},
		Auth: config.AuthConfig{
			JWTSecret:     "test-jwt-secret-1234567890",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: -1 * time.Minute,
		},
	}
	expManager := NewTokenManager(expCfg)
	expiredToken, _, err := expManager.IssueRefreshToken("user-1")
	if err != nil {
		t.Fatalf("issue expired token: %v", err)
	}
	if _, err := manager.ParseRefreshToken(expiredToken); err == nil {
		t.Fatal("expected parse refresh token to reject expired token")
	}
}

func TestTokenManager_ParseAccessToken_MalformedAndInvalid(t *testing.T) {
	manager := newTestTokenManager()

	cases := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "malformed text", token: "invalid.jwt.token"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := manager.ParseAccessToken(tc.token); err == nil {
				t.Fatalf("expected parse access token error for %q", tc.name)
			}
		})
	}

	// Expired access token
	expCfg := config.Config{
		Application: config.ApplicationConfig{AppName: "recova-auth-test", AppURL: "http://localhost:3000"},
		Auth: config.AuthConfig{
			JWTSecret:     "test-jwt-secret-1234567890",
			JWTAccessTTL:  -1 * time.Minute,
			JWTRefreshTTL: 24 * time.Hour,
		},
	}
	expManager := NewTokenManager(expCfg)
	expiredToken, _, err := expManager.IssueAccessToken("user-1")
	if err != nil {
		t.Fatalf("issue expired access token: %v", err)
	}
	if _, err := manager.ParseAccessToken(expiredToken); err == nil {
		t.Fatal("expected parse access token to reject expired token")
	}
}

func newTestTokenManager() *TokenManager {
	cfg := config.Config{
		Application: config.ApplicationConfig{
			AppName: "recova-auth-test",
			AppURL:  "http://localhost:3000",
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
