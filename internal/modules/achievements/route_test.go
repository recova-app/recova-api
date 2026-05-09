package achievements

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
	authService := buildAchievementsAuthService(t, "user-1")
	service := NewService(&achievementsRouteRepo{})

	app := newAchievementsTestApp()
	RegisterRoutes(app.Group("/api/v1/achievements"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/achievements/catalog", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_Success(t *testing.T) {
	authService := buildAchievementsAuthService(t, "user-1")
	service := NewService(&achievementsRouteRepo{})
	service.now = func() time.Time {
		return time.Date(2026, 5, 9, 7, 0, 0, 0, time.UTC)
	}

	app := newAchievementsTestApp()
	RegisterRoutes(app.Group("/api/v1/achievements"), authService, service)

	tests := []struct {
		name string
		path string
	}{
		{name: "catalog", path: "/api/v1/achievements/catalog"},
		{name: "progress", path: "/api/v1/achievements/progress"},
		{name: "unlocked", path: "/api/v1/achievements/unlocked"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := httpharness.JSONRequest(t, app, fiber.MethodGet, tc.path, nil, map[string]string{
				"Authorization": "Bearer access-token",
			})
			httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
			httpharness.RequireSuccessEnvelope(t, resp.JSON)
		})
	}
}

type achievementsRouteRepo struct{}

func (r *achievementsRouteRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"}, nil
}

func (r *achievementsRouteRepo) ListActiveAchievements(_ context.Context, _ *string) ([]models.Achievement, error) {
	return []models.Achievement{{
		ID:          "a-1",
		Code:        "streak_7_days",
		Title:       "7 Hari",
		Description: "desc",
		Category:    categoryStreakMilestone,
		Threshold:   7,
		IsActive:    true,
	}}, nil
}

func (r *achievementsRouteRepo) ListProgressByUser(_ context.Context, _ string, _ *string) ([]progressListRow, error) {
	return []progressListRow{{
		Code:          "streak_7_days",
		Category:      categoryStreakMilestone,
		Threshold:     7,
		ProgressValue: 8,
		UnlockedAt:    ptrTime(time.Date(2026, 5, 9, 7, 0, 0, 0, time.UTC)),
	}}, nil
}

func (r *achievementsRouteRepo) ListUnlockedByUser(_ context.Context, _ string, _ *string) ([]unlockedListRow, error) {
	return []unlockedListRow{{
		Code:          "streak_7_days",
		Title:         "7 Hari",
		Description:   "desc",
		Category:      categoryStreakMilestone,
		Threshold:     7,
		ProgressValue: 8,
		UnlockedAt:    time.Date(2026, 5, 9, 7, 0, 0, 0, time.UTC),
	}}, nil
}

func (r *achievementsRouteRepo) ComputeEvaluationMetrics(_ context.Context, _ string, _ time.Time) (evaluationMetrics, error) {
	return evaluationMetrics{StreakMilestone: 8}, nil
}

func (r *achievementsRouteRepo) UpsertProgress(_ context.Context, _ string, _ []progressUpsert, _ time.Time) error {
	return nil
}

func buildAchievementsAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &achievementsAuthRepo{user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"}}
	tokens := &achievementsAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		}},
	}
	return authmodule.NewService(repo, &achievementsAuthVerifier{}, tokens)
}

type achievementsAuthRepo struct {
	user models.User
}

func (r *achievementsAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *achievementsAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *achievementsAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *achievementsAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *achievementsAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *achievementsAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *achievementsAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *achievementsAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type achievementsAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *achievementsAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *achievementsAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *achievementsAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}, nil
}
func (p *achievementsAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *achievementsAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}, nil
}
func (p *achievementsAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *achievementsAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *achievementsAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *achievementsAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string {
	return "refresh-token"
}

type achievementsAuthVerifier struct{}

func (v *achievementsAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newAchievementsTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
