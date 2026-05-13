package content

import (
	"context"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
)

func TestService_GetDailyContent_DeterministicSelection(t *testing.T) {
	repo := &fakeContentRepo{
		motivations: []models.DailyMotivation{
			{Content: "motivasi-a"},
			{Content: "motivasi-b"},
		},
		challenges: []models.DailyChallenge{
			{Content: "challenge-a"},
			{Content: "challenge-b"},
		},
	}

	service := NewService(repo)
	service.now = func() time.Time {
		return time.Date(2026, 5, 8, 13, 0, 0, 0, time.UTC)
	}

	first, err := service.GetDailyContent(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := service.GetDailyContent(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != second {
		t.Fatalf("expected deterministic payload, got first=%+v second=%+v", first, second)
	}
}

func TestService_GetDailyContent_FallbackWhenEmpty(t *testing.T) {
	repo := &fakeContentRepo{}
	service := NewService(repo)
	service.now = func() time.Time {
		return time.Date(2026, 5, 8, 13, 0, 0, 0, time.UTC)
	}

	payload, err := service.GetDailyContent(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Motivation != fallbackMotivation {
		t.Fatalf("expected fallback motivation, got %q", payload.Motivation)
	}
	if payload.Challenge != fallbackChallenge {
		t.Fatalf("expected fallback challenge, got %q", payload.Challenge)
	}
}

type fakeContentRepo struct {
	motivations []models.DailyMotivation
	challenges  []models.DailyChallenge
}

func (r *fakeContentRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"}, nil
}

func (r *fakeContentRepo) ListActiveMotivations(_ context.Context) ([]models.DailyMotivation, error) {
	return r.motivations, nil
}

func (r *fakeContentRepo) ListActiveChallenges(_ context.Context) ([]models.DailyChallenge, error) {
	return r.challenges, nil
}
