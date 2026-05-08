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

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "Halo"})
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
		Message:  "payload rejected: user prompt 'sangat sensitif'",
	}}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "sangat sensitif"})
	if err == nil {
		t.Fatal("expected error")
	}

	mapped := errs.Map(err)
	if mapped.Code != errs.CodeDownstreamError {
		t.Fatalf("expected DOWNSTREAM_ERROR, got %s", mapped.Code)
	}
	if strings.Contains(strings.ToLower(mapped.Message), "sangat sensitif") {
		t.Fatalf("expected safe message, got %q", mapped.Message)
	}
}

func TestService_AskCoach_SuccessStoresConversation(t *testing.T) {
	profileReason := "ingin sehat"
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test", UserWhy: &profileReason},
		history: []models.AIChat{{Role: "user", Content: "pesan lama"}},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Tetap semangat"}}
	service := NewService(repo, provider)
	service.now = func() time.Time {
		return time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC)
	}

	payload, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "Hari ini berat"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Response != "Tetap semangat" {
		t.Fatalf("unexpected response: %q", payload.Response)
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
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "bukan-json"}}
	service := NewService(repo, provider)

	_, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frekuensi": "harian"},
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
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: `{"level":"Sedang","title":"Analisis Awal","level_description":"Penjelasan","pattern_analysis":"Pola stres","encouragement":"Kamu bisa"}`}}
	service := NewService(repo, provider)

	payload, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frekuensi": "harian"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Level != "Sedang" || payload.PatternAnalysis != "Pola stres" {
		t.Fatalf("unexpected payload: %+v", payload)
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

type fakeAIProvider struct {
	response aiplatform.GenerateResponse
	err      error
}

func (p *fakeAIProvider) Generate(_ context.Context, _ aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	if p.err != nil {
		return aiplatform.GenerateResponse{}, p.err
	}
	if strings.TrimSpace(p.response.Text) == "" {
		return aiplatform.GenerateResponse{}, errors.New("missing response")
	}
	return p.response, nil
}
