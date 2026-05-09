package routine

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
	authService := buildRoutineAuthService(t, "user-1")
	service := NewService(&fakeRoutineRepo{})

	app := newRoutineTestApp()
	RegisterRoutes(app.Group("/api/v1/routine"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/routine/statistics", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_CheckInValidationError(t *testing.T) {
	authService := buildRoutineAuthService(t, "user-1")
	service := NewService(&fakeRoutineRepo{})

	app := newRoutineTestApp()
	RegisterRoutes(app.Group("/api/v1/routine"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/routine/checkin", map[string]any{
		"mood": "",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterRoutes_GetStatisticsSuccess(t *testing.T) {
	authService := buildRoutineAuthService(t, "user-1")
	service := NewService(&fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		successfulRows: []models.CheckIn{
			{CheckInDate: time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC)},
		},
	})
	service.now = func() time.Time { return time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC) }

	app := newRoutineTestApp()
	RegisterRoutes(app.Group("/api/v1/routine"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/routine/statistics", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)

	data, ok := resp.JSON["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp.JSON["data"])
	}
	for _, key := range []string{
		"currentStreak",
		"longestStreak",
		"totalCheckins",
		"streakCalendar",
		"relapseCount",
		"relapseRate",
		"recoverySuccessRate",
		"checkinConsistencyScore",
		"weeklyProgress",
		"monthlyProgress",
		"moodTrend",
	} {
		if _, exists := data[key]; !exists {
			t.Fatalf("expected statistics field %s", key)
		}
	}
}

func TestRegisterRoutes_GetActivitySummaryValidationError(t *testing.T) {
	authService := buildRoutineAuthService(t, "user-1")
	service := NewService(&fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
	})

	app := newRoutineTestApp()
	RegisterRoutes(app.Group("/api/v1/routine"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/routine/statistics/activity-summary?windowDays=3", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterRoutes_GetActivitySummarySuccess(t *testing.T) {
	authService := buildRoutineAuthService(t, "user-1")
	service := NewService(&fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		windowRows: []models.CheckIn{
			{
				ID:           "checkin-1",
				CheckInDate:  time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
				Mood:         "tenang",
				IsSuccessful: true,
				CreatedAt:    time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC),
			},
		},
	})
	service.now = func() time.Time { return time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC) }

	app := newRoutineTestApp()
	RegisterRoutes(app.Group("/api/v1/routine"), authService, service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/routine/statistics/activity-summary?windowDays=30", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)

	data, ok := resp.JSON["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", resp.JSON["data"])
	}
	for _, key := range []string{
		"windowDays",
		"successfulCheckins",
		"relapses",
		"activeDays",
		"recentActivity",
	} {
		if _, exists := data[key]; !exists {
			t.Fatalf("expected activity summary field %s", key)
		}
	}
}

func buildRoutineAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &routineAuthRepo{
		user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"},
	}
	tokens := &routineAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		}},
	}
	return authmodule.NewService(repo, &routineAuthVerifier{}, tokens)
}

type routineAuthRepo struct {
	user models.User
}

func (r *routineAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *routineAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *routineAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *routineAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *routineAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *routineAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *routineAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *routineAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type routineAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *routineAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *routineAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *routineAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}, nil
}
func (p *routineAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *routineAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}, nil
}
func (p *routineAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *routineAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *routineAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *routineAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type routineAuthVerifier struct{}

func (v *routineAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newRoutineTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
