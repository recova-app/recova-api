package journals

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
	authService := buildJournalsAuthService(t, "user-1")
	service := NewService(&fakeJournalsRepo{})

	app := newJournalsTestApp()
	RegisterRoutes(app.Group("/api/v1/journals"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/journals", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_CreateValidationError(t *testing.T) {
	authService := buildJournalsAuthService(t, "user-1")
	service := NewService(&fakeJournalsRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
	})

	app := newJournalsTestApp()
	RegisterRoutes(app.Group("/api/v1/journals"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/journals", map[string]any{
		"content": "",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterRoutes_ListSuccess(t *testing.T) {
	authService := buildJournalsAuthService(t, "user-1")
	service := NewService(&fakeJournalsRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		listRows: []models.Journal{
			{ID: "journal-1", UserID: "user-1", Content: "isi", CreatedAt: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)},
		},
	})

	app := newJournalsTestApp()
	RegisterRoutes(app.Group("/api/v1/journals"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/journals", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func buildJournalsAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &journalsAuthRepo{
		user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"},
	}
	tokens := &journalsAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		}},
	}
	return authmodule.NewService(repo, &journalsAuthVerifier{}, tokens)
}

type journalsAuthRepo struct {
	user models.User
}

func (r *journalsAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *journalsAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *journalsAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *journalsAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *journalsAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *journalsAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *journalsAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *journalsAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type journalsAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *journalsAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *journalsAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *journalsAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}, nil
}
func (p *journalsAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *journalsAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}, nil
}
func (p *journalsAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *journalsAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *journalsAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *journalsAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type journalsAuthVerifier struct{}

func (v *journalsAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newJournalsTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
