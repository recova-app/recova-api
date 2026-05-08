package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/recova-app/backend-v2/internal/platform/config"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

// SessionClaims represents signed JWT claims for Recova sessions.
type SessionClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenManager creates and validates access/refresh tokens.
type TokenManager struct {
	secret         []byte
	issuer         string
	accessTTL      time.Duration
	refreshTTL     time.Duration
	cookieName     string
	cookieSecure   bool
	cookieSameSite string
	cookieDomain   string
	googleClientID string
}

// NewTokenManager builds token manager from validated runtime config.
func NewTokenManager(cfg config.Config) *TokenManager {
	return &TokenManager{
		secret:         []byte(cfg.Auth.JWTSecret),
		issuer:         strings.TrimSpace(cfg.Application.AppName),
		accessTTL:      cfg.Auth.JWTAccessTTL,
		refreshTTL:     cfg.Auth.JWTRefreshTTL,
		cookieName:     strings.TrimSpace(cfg.Auth.Cookie.Name),
		cookieSecure:   cfg.Auth.Cookie.Secure,
		cookieSameSite: strings.TrimSpace(cfg.Auth.Cookie.SameSite),
		cookieDomain:   strings.TrimSpace(cfg.Auth.Cookie.Domain),
		googleClientID: strings.TrimSpace(cfg.Auth.GoogleClient),
	}
}

// GoogleAudience returns configured Google OAuth audience.
func (m *TokenManager) GoogleAudience() string {
	return m.googleClientID
}

// IssueAccessToken creates signed short-lived access token.
func (m *TokenManager) IssueAccessToken(userID string) (string, SessionPayload, error) {
	claims := m.newClaims(userID, tokenTypeAccess, m.accessTTL)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", SessionPayload{}, fmt.Errorf("sign access token: %w", err)
	}

	return token, SessionPayload{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(m.accessTTL.Seconds()),
	}, nil
}

// IssueRefreshToken creates signed refresh token and returns raw token plus claims id/exp.
func (m *TokenManager) IssueRefreshToken(userID string) (string, SessionClaims, error) {
	claims := m.newClaims(userID, tokenTypeRefresh, m.refreshTTL)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", SessionClaims{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return token, claims, nil
}

// ParseAccessToken validates access token and returns claims.
func (m *TokenManager) ParseAccessToken(rawToken string) (SessionClaims, error) {
	return m.parseToken(rawToken, tokenTypeAccess)
}

// ParseRefreshToken validates refresh token and returns claims.
func (m *TokenManager) ParseRefreshToken(rawToken string) (SessionClaims, error) {
	return m.parseToken(rawToken, tokenTypeRefresh)
}

// HashRefreshToken computes deterministic hash of refresh token for storage/revocation.
func (m *TokenManager) HashRefreshToken(rawToken string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(rawToken)))
	return hex.EncodeToString(digest[:])
}

// RefreshCookie builds httpOnly refresh cookie value.
func (m *TokenManager) RefreshCookie(rawToken string) *fiber.Cookie {
	cookie := &fiber.Cookie{
		Name:     m.cookieName,
		Value:    strings.TrimSpace(rawToken),
		Path:     "/",
		HTTPOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.cookieSameSite,
		Expires:  time.Now().Add(m.refreshTTL),
	}
	if m.cookieDomain != "" {
		cookie.Domain = m.cookieDomain
	}
	return cookie
}

// ExpiredRefreshCookie builds immediate-expire refresh cookie for logout flow.
func (m *TokenManager) ExpiredRefreshCookie() *fiber.Cookie {
	cookie := &fiber.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.cookieSameSite,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	}
	if m.cookieDomain != "" {
		cookie.Domain = m.cookieDomain
	}
	return cookie
}

// RefreshCookieValue extracts refresh token from request cookies.
func (m *TokenManager) RefreshCookieValue(c fiber.Ctx) string {
	return strings.TrimSpace(c.Cookies(m.cookieName))
}

func (m *TokenManager) newClaims(userID string, tokenType string, ttl time.Duration) SessionClaims {
	now := time.Now().UTC()
	return SessionClaims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strings.TrimSpace(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
}

func (m *TokenManager) parseToken(rawToken string, expectedType string) (SessionClaims, error) {
	raw := strings.TrimSpace(rawToken)
	if raw == "" {
		return SessionClaims{}, fmt.Errorf("token is required")
	}

	claims := &SessionClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unsupported token algorithm")
		}
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return SessionClaims{}, err
	}

	if strings.TrimSpace(claims.TokenType) != expectedType {
		return SessionClaims{}, fmt.Errorf("invalid token type")
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return SessionClaims{}, fmt.Errorf("token subject is required")
	}

	return *claims, nil
}
