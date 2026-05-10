package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

func TestService_AskCoach_ProviderTimeoutMapsServiceUnavailable(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	provider := &fakeAIProvider{err: &aiplatform.ProviderError{Provider: aiplatform.ProviderGemini, Kind: aiplatform.ErrorKindTimeout, Message: "timeout"}}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "Hello"})
	if err == nil {
		t.Fatal("expected error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeServiceUnavailable {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %s", mapped.Code)
	}
}

func TestService_AskCoach_ProviderErrorMessageIsSafe(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	provider := &fakeAIProvider{err: &aiplatform.ProviderError{
		Provider: aiplatform.ProviderOpenAICompatible,
		Kind:     aiplatform.ErrorKindInvalidResponse,
		Message:  "payload rejected: user prompt 'highly sensitive'",
	}}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "highly sensitive"})
	if err == nil {
		t.Fatal("expected error")
	}

	mapped := errs.Map(err)
	if mapped.Code != errs.CodeDownstreamError {
		t.Fatalf("expected DOWNSTREAM_ERROR, got %s", mapped.Code)
	}
	if strings.Contains(strings.ToLower(mapped.Message), "highly sensitive") {
		t.Fatalf("expected safe message, got %q", mapped.Message)
	}
}

func TestService_AskCoach_SuccessStoresConversation(t *testing.T) {
	profileReason := "want to be healthy"
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test", UserWhy: &profileReason},
		history: []models.AIChat{{Role: "user", Content: "old message"}},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Stay strong"}}
	service := NewService(repo, provider)
	service.now = func() time.Time {
		return time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC)
	}

	payload, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "Today is hard"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Response != "Stay strong" {
		t.Fatalf("unexpected response: %q", payload.Response)
	}
	if payload.PersonaUsed != DefaultPersona {
		t.Fatalf("expected default persona %q, got %q", DefaultPersona, payload.PersonaUsed)
	}
	if len(repo.createdMessages) != 2 {
		t.Fatalf("expected 2 stored rows, got %d", len(repo.createdMessages))
	}
	if repo.createdMessages[0].Role != "user" || repo.createdMessages[1].Role != "model" {
		t.Fatalf("unexpected stored roles: %+v", repo.createdMessages)
	}
}

func TestService_GetSummary_FallbackWhenNoProfile(t *testing.T) {
	repo := &fakeAIRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		profileErr: gorm.ErrRecordNotFound,
	}
	service := NewService(repo, &fakeAIProvider{})

	payload, err := service.GetSummary(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Summary != defaultSummaryMessage {
		t.Fatalf("expected fallback summary, got %q", payload.Summary)
	}
}

func TestService_AnalyzeOnboarding_InvalidProviderJSON(t *testing.T) {
	repo := &fakeAIRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "not-json"}}
	service := NewService(repo, provider)

	_, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frequency": "daily"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeDownstreamError {
		t.Fatalf("expected DOWNSTREAM_ERROR, got %s", mapped.Code)
	}
}

func TestService_AnalyzeOnboarding_Success(t *testing.T) {
	repo := &fakeAIRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: `{"level":"Moderate","title":"Initial Analysis","level_description":"Explanation","pattern_analysis":"Stress pattern","encouragement":"You can do this"}`}}
	service := NewService(repo, provider)

	payload, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frequency": "daily"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Level != "Moderate" || payload.PatternAnalysis != "Stress pattern" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestService_GetPersonaPreference_FallbackDefault(t *testing.T) {
	repo := &fakeAIRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		personaErr: gorm.ErrRecordNotFound,
	}
	service := NewService(repo, &fakeAIProvider{})

	payload, err := service.GetPersonaPreference(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Persona != DefaultPersona {
		t.Fatalf("expected default persona %q, got %q", DefaultPersona, payload.Persona)
	}
	if payload.FallbackPersona != DefaultPersona {
		t.Fatalf("unexpected fallback persona: %q", payload.FallbackPersona)
	}
}

func TestService_UpdatePersonaPreference_Validation(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	service := NewService(repo, &fakeAIProvider{})

	_, err := service.UpdatePersonaPreference(context.Background(), "user-1", PersonaPreferenceRequest{Persona: "invalid"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR, got %s", mapped.Code)
	}
}

func TestService_AskCoach_UsesStoredPersonaInProviderRequest(t *testing.T) {
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		persona: models.UserAIPersonaPreference{UserID: "user-1", Persona: "friendly"},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "ok"}}
	service := NewService(repo, provider)

	payload, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.PersonaUsed != "friendly" {
		t.Fatalf("expected persona_used friendly, got %q", payload.PersonaUsed)
	}
	if provider.lastReq.Persona != "friendly" {
		t.Fatalf("expected provider persona friendly, got %q", provider.lastReq.Persona)
	}
	if !strings.Contains(provider.lastReq.SystemInstruction, "persona name: friendly") {
		t.Fatalf("expected persona marker in system instruction, got: %s", provider.lastReq.SystemInstruction)
	}
}

type fakeAIRepo struct {
	user            models.User
	userErr         error
	profile         models.Profile
	profileErr      error
	activeStreak    *models.Streak
	activeStreakErr error
	history         []models.AIChat
	historyErr      error
	createdMessages []models.AIChat
	createErr       error
	persona         models.UserAIPersonaPreference
	personaErr      error
	upsertErr       error
	upsertPersona   models.UserAIPersonaPreference
}

func (r *fakeAIRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.userErr != nil {
		return models.User{}, r.userErr
	}
	if strings.TrimSpace(r.user.ID) == "" {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *fakeAIRepo) FindProfileByUserID(_ context.Context, _ string) (models.Profile, error) {
	if r.profileErr != nil {
		return models.Profile{}, r.profileErr
	}
	return r.profile, nil
}

func (r *fakeAIRepo) FindActiveStreakByUserID(_ context.Context, _ string) (*models.Streak, error) {
	if r.activeStreakErr != nil {
		return nil, r.activeStreakErr
	}
	return r.activeStreak, nil
}

func (r *fakeAIRepo) ListRecentChatsByUserID(_ context.Context, _ string, _ int) ([]models.AIChat, error) {
	if r.historyErr != nil {
		return nil, r.historyErr
	}
	return r.history, nil
}

func (r *fakeAIRepo) CreateChatMessages(_ context.Context, rows []models.AIChat) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.createdMessages = rows
	return nil
}

func (r *fakeAIRepo) GetPersonaPreferenceByUserID(_ context.Context, _ string) (models.UserAIPersonaPreference, error) {
	if r.personaErr != nil {
		return models.UserAIPersonaPreference{}, r.personaErr
	}
	if strings.TrimSpace(r.persona.UserID) == "" {
		return models.UserAIPersonaPreference{}, gorm.ErrRecordNotFound
	}
	return r.persona, nil
}

func (r *fakeAIRepo) UpsertPersonaPreference(_ context.Context, userID string, persona string, updatedAt time.Time) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upsertPersona = models.UserAIPersonaPreference{
		UserID:    strings.TrimSpace(userID),
		Persona:   strings.TrimSpace(persona),
		UpdatedAt: updatedAt,
	}
	return nil
}

type fakeAIProvider struct {
	response aiplatform.GenerateResponse
	err      error
	lastReq  aiplatform.GenerateRequest
}

func (p *fakeAIProvider) Generate(_ context.Context, req aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	p.lastReq = req
	if p.err != nil {
		return aiplatform.GenerateResponse{}, p.err
	}
	if strings.TrimSpace(p.response.Text) == "" {
		return aiplatform.GenerateResponse{}, errors.New("missing response")
	}
	return p.response, nil
}
