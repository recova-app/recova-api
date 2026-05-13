package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
	"gorm.io/gorm"
)

func TestRegisterRoutes_Unauthenticated(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}, &aiRouteProvider{response: aiplatform.GenerateResponse{Text: "ok"}})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/ai/ask-coach", map[string]any{"message": "Halo"}, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_AskCoachSuccess(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}, &aiRouteProvider{response: aiplatform.GenerateResponse{Text: "Tetap semangat"}})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/ai/ask-coach", map[string]any{"message": "Halo"}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_AskCoachProviderTimeoutSafeError(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}, &aiRouteProvider{
		err: &aiplatform.ProviderError{Provider: aiplatform.ProviderGemini, Kind: aiplatform.ErrorKindTimeout, Message: "timeout"},
	})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/ai/ask-coach", map[string]any{"message": "Halo"}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusServiceUnavailable)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "SERVICE_UNAVAILABLE")
}

func TestRegisterRoutes_OnboardingAnalysisValidation(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}, &aiRouteProvider{response: aiplatform.GenerateResponse{Text: "{}"}})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/ai/onboarding-analysis", map[string]any{"answers": map[string]any{}}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterRoutes_ChatHistorySuccess(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		history: []models.AIChat{{ID: "chat-1", Role: "user", Content: "hai", CreatedAt: time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC)}},
	}, &aiRouteProvider{response: aiplatform.GenerateResponse{Text: "ok"}})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/ai/chat-history", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_RelapseSolutionSuccess(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}, &aiRouteProvider{
		response: aiplatform.GenerateResponse{Text: `{"title":"Pulihkan Fokus","analysis":"Relapse terjadi saat stres malam.","action_steps":["Minum air","Tutup aplikasi pemicu","Hubungi teman support"]}`},
	})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/ai/relapse-solution", map[string]any{
		"mood":            "cemas",
		"relapse_trigger": []string{"stres kerja", "sendiri malam"},
		"commitment":      "ingin kembali stabil",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_GetPersonaPreferenceSuccess(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		personaRow: models.UserAIPersonaPreference{UserID: "user-1", Persona: "friendly"},
	}, &aiRouteProvider{response: aiplatform.GenerateResponse{Text: "ok"}})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/ai/persona-preferences", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_UpdatePersonaPreferenceValidation(t *testing.T) {
	authService := buildAIAuthService(t, "user-1")
	service := NewService(&aiRouteRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}, &aiRouteProvider{response: aiplatform.GenerateResponse{Text: "ok"}})

	app := newAITestApp()
	RegisterRoutes(app.Group("/api/v1/ai"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPut, "/api/v1/ai/persona-preferences", map[string]any{
		"persona": "unknown",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func buildAIAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &aiAuthRepo{user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"}}
	tokens := &aiAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: userID, ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute))}},
	}
	return authmodule.NewService(repo, &aiAuthVerifier{}, tokens)
}

type aiRouteRepo struct {
	user              models.User
	history           []models.AIChat
	personaRow        models.UserAIPersonaPreference
	personaErr        error
	upsertPreference  models.UserAIPersonaPreference
	upsertPreferenceE error
}

func (r *aiRouteRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.user.ID == "" {
		return models.User{}, errors.New("user not found")
	}
	return r.user, nil
}

func (r *aiRouteRepo) FindProfileByUserID(_ context.Context, _ string) (models.Profile, error) {
	return models.Profile{}, nil
}

func (r *aiRouteRepo) FindActiveStreakByUserID(_ context.Context, _ string) (*models.Streak, error) {
	return nil, nil
}

func (r *aiRouteRepo) ListRecentChatsByUserID(_ context.Context, _ string, _ int) ([]models.AIChat, error) {
	return r.history, nil
}

func (r *aiRouteRepo) CreateChatMessages(_ context.Context, _ []models.AIChat) error {
	return nil
}

func (r *aiRouteRepo) GetPersonaPreferenceByUserID(_ context.Context, _ string) (models.UserAIPersonaPreference, error) {
	if r.personaErr != nil {
		return models.UserAIPersonaPreference{}, r.personaErr
	}
	if r.personaRow.UserID == "" {
		return models.UserAIPersonaPreference{}, gorm.ErrRecordNotFound
	}
	return r.personaRow, nil
}

func (r *aiRouteRepo) UpsertPersonaPreference(_ context.Context, userID string, persona string, updatedAt time.Time) error {
	if r.upsertPreferenceE != nil {
		return r.upsertPreferenceE
	}
	r.upsertPreference = models.UserAIPersonaPreference{
		UserID:    userID,
		Persona:   persona,
		UpdatedAt: updatedAt,
	}
	return nil
}

type aiRouteProvider struct {
	response aiplatform.GenerateResponse
	err      error
}

func (p *aiRouteProvider) Generate(_ context.Context, _ aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	if p.err != nil {
		return aiplatform.GenerateResponse{}, p.err
	}
	return p.response, nil
}

type aiAuthRepo struct {
	user models.User
}

func (r *aiAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *aiAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *aiAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *aiAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *aiAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *aiAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *aiAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *aiAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type aiAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *aiAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *aiAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *aiAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *aiAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *aiAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *aiAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *aiAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *aiAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *aiAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type aiAuthVerifier struct{}

func (v *aiAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newAITestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
