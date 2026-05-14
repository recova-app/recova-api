package education

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
)

func TestRegisterRoutes_Unauthenticated(t *testing.T) {
	authService := buildEducationAuthService(t, "user-1")
	service := NewService(&educationRouteRepo{})

	app := newEducationTestApp()
	RegisterRoutes(app.Group("/api/v1/education"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/education", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_Success(t *testing.T) {
	authService := buildEducationAuthService(t, "user-1")
	service := NewService(&educationRouteRepo{})

	app := newEducationTestApp()
	RegisterRoutes(app.Group("/api/v1/education"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/education", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)

	data, ok := resp.JSON["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("unexpected data payload: %#v", resp.JSON["data"])
	}
	first, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first item payload: %#v", data[0])
	}
	contentType, ok := first["type"].(string)
	if !ok || contentType != "artikel" {
		t.Fatalf("unexpected type payload: %#v", first["type"])
	}
	category, ok := first["category"].(string)
	if !ok {
		t.Fatalf("unexpected category payload: %#v", first["category"])
	}
	if strings.Contains(category, "_") {
		t.Fatalf("category must not contain underscore: %s", category)
	}
}

type educationRouteRepo struct{}

func (r *educationRouteRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"}, nil
}

func (r *educationRouteRepo) ListActiveContents(_ context.Context) ([]models.EducationContent, error) {
	published_at := time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	return []models.EducationContent{{
		ID:          "education-1",
		Title:       "judul",
		Description: ptrStringEducation("deskripsi"),
		URL:         "https://example.test/edu",
		Category:    "regulasi_emosi",
		Type:        "artikel",
		PublishedAt: &published_at,
	}}, nil
}

func ptrStringEducation(value string) *string {
	return &value
}

func buildEducationAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &educationAuthRepo{user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"}}
	tokens := &educationAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: userID, ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute))}},
	}
	return authmodule.NewService(repo, &educationAuthVerifier{}, tokens)
}

type educationAuthRepo struct {
	user models.User
}

func (r *educationAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *educationAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *educationAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *educationAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *educationAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *educationAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *educationAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *educationAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type educationAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *educationAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *educationAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *educationAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *educationAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *educationAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *educationAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *educationAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *educationAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *educationAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type educationAuthVerifier struct{}

func (v *educationAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newEducationTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
