package users

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
	"gorm.io/gorm"
)

func TestRegisterUserRoutes_Unauthenticated(t *testing.T) {
	authService := buildUsersAuthService(t, "user-1")
	usersService := NewService(&usersRouteRepo{}, "local", "development")

	app := newUsersTestApp()
	RegisterUserRoutes(app.Group("/api/v1/users"), authService, usersService)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/users/me", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterOnboardingRoute_ValidationError(t *testing.T) {
	authService := buildUsersAuthService(t, "user-1")
	usersService := NewService(&usersRouteRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
	}, "local", "development")

	app := newUsersTestApp()
	RegisterOnboardingRoute(app.Group("/api/v1/auth"), authService, usersService)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/onboarding", map[string]any{}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterUserRoutes_GetMeSuccess(t *testing.T) {
	reason := "Fokus"
	checkIn := "09:00"
	authService := buildUsersAuthService(t, "user-1")
	usersService := NewService(&usersRouteRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester", UserWhy: &reason, CheckInTime: &checkIn},
	}, "local", "development")

	app := newUsersTestApp()
	RegisterUserRoutes(app.Group("/api/v1/users"), authService, usersService)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/users/me", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func buildUsersAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &usersAuthRepo{
		user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"},
	}
	tokens := &usersAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: userID, ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute))}},
	}
	return authmodule.NewService(repo, &usersAuthVerifier{}, tokens)
}

type usersRouteRepo struct {
	user models.User
}

func (r *usersRouteRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.user.ID == "" {
		return models.User{}, errors.New("missing user")
	}
	return r.user, nil
}

func (r *usersRouteRepo) FindProfileByUserID(_ context.Context, _ string) (models.Profile, error) {
	return models.Profile{}, gorm.ErrRecordNotFound
}

func (r *usersRouteRepo) UpdateUserFields(_ context.Context, _ string, _ map[string]any) error {
	return nil
}

func (r *usersRouteRepo) CompleteOnboarding(_ context.Context, _ string, _ OnboardingInput) (models.User, models.Profile, error) {
	return r.user, models.Profile{ID: "profile-1", UserID: r.user.ID}, nil
}

func (r *usersRouteRepo) ResetUserDataForTesting(_ context.Context, _ string) error {
	return nil
}

type usersAuthRepo struct {
	user models.User
}

func (r *usersAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *usersAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *usersAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *usersAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *usersAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *usersAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *usersAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *usersAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type usersAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *usersAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *usersAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *usersAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *usersAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *usersAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *usersAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *usersAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *usersAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *usersAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type usersAuthVerifier struct{}

func (v *usersAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, nil
}

func newUsersTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
