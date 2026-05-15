package content

import (
	"context"
	"errors"
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
	authService := buildContentAuthService(t, "user-1")
	service := NewService(&contentRouteRepo{})

	app := newContentTestApp()
	RegisterRoutes(app.Group("/api/v1/content"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/content/daily", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_Success(t *testing.T) {
	authService := buildContentAuthService(t, "user-1")
	service := NewService(&contentRouteRepo{})

	app := newContentTestApp()
	RegisterRoutes(app.Group("/api/v1/content"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/content/daily", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

type contentRouteRepo struct{}

func (r *contentRouteRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"}, nil
}

func (r *contentRouteRepo) ListActiveMotivations(_ context.Context) ([]models.DailyMotivation, error) {
	return []models.DailyMotivation{{Content: "motivasi-a"}}, nil
}

func (r *contentRouteRepo) ListActiveChallenges(_ context.Context) ([]models.DailyChallenge, error) {
	return []models.DailyChallenge{{
		Title:       "challenge-title-a",
		Description: "challenge-description-a",
		Content:     "challenge-content-a",
	}}, nil
}

func (r *contentRouteRepo) ListActivePhysicalChallenges(_ context.Context) ([]models.DailyPhysicalChallenge, error) {
	return []models.DailyPhysicalChallenge{{
		Title:       "physical-title-a",
		Description: "physical-description-a",
	}}, nil
}

func buildContentAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &contentAuthRepo{user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"}}
	tokens := &contentAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: userID, ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute))}},
	}
	return authmodule.NewService(repo, &contentAuthVerifier{}, tokens)
}

type contentAuthRepo struct {
	user models.User
}

func (r *contentAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *contentAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *contentAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *contentAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *contentAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *contentAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *contentAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *contentAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type contentAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *contentAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *contentAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *contentAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *contentAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *contentAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *contentAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *contentAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *contentAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *contentAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type contentAuthVerifier struct{}

func (v *contentAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newContentTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
