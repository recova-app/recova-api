package achievements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

func TestService_GetCatalog_Success(t *testing.T) {
	repo := &fakeAchievementsRepo{
		user: models.User{ID: "user-1"},
		achievements: []models.Achievement{
			{ID: "a-1", Code: "streak_7_days", Title: "7 Hari", Description: "desc", Category: categoryStreakMilestone, Threshold: 7, IsActive: true},
		},
	}
	service := NewService(repo)

	payload, err := service.GetCatalog(context.Background(), "user-1", CategoryQuery{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}
	if payload.Items[0].Code != "streak_7_days" {
		t.Fatalf("unexpected code: %s", payload.Items[0].Code)
	}
}

func TestService_GetProgress_EvaluateAndPersist(t *testing.T) {
	now := time.Date(2026, 5, 9, 7, 0, 0, 0, time.UTC)
	repo := &fakeAchievementsRepo{
		user: models.User{ID: "user-1"},
		achievements: []models.Achievement{
			{ID: "a-1", Code: "streak_7_days", Category: categoryStreakMilestone, Threshold: 7, IsActive: true},
			{ID: "a-2", Code: "onboarding_complete", Category: categoryOnboardingCompletion, Threshold: 1, IsActive: true},
		},
		metrics: evaluationMetrics{
			StreakMilestone:      8,
			OnboardingCompletion: 1,
		},
		progressRows: []progressListRow{
			{Code: "streak_7_days", Category: categoryStreakMilestone, Threshold: 7, ProgressValue: 8, UnlockedAt: &now},
			{Code: "onboarding_complete", Category: categoryOnboardingCompletion, Threshold: 1, ProgressValue: 1, UnlockedAt: &now},
		},
	}
	service := NewService(repo)
	service.now = func() time.Time { return now }

	payload, err := service.GetProgress(context.Background(), "user-1", CategoryQuery{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(repo.upsertRows) != 2 {
		t.Fatalf("expected 2 upsert rows, got %d", len(repo.upsertRows))
	}
	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 progress items, got %d", len(payload.Items))
	}
	if !payload.Items[0].Unlocked {
		t.Fatal("expected first progress unlocked")
	}
}

func TestService_GetUnlocked_Success(t *testing.T) {
	now := time.Date(2026, 5, 9, 7, 0, 0, 0, time.UTC)
	repo := &fakeAchievementsRepo{
		user: models.User{ID: "user-1"},
		achievements: []models.Achievement{
			{ID: "a-1", Code: "streak_7_days", Category: categoryStreakMilestone, Threshold: 7, IsActive: true},
		},
		metrics: evaluationMetrics{StreakMilestone: 8},
		unlockedRows: []unlockedListRow{
			{
				Code:          "streak_7_days",
				Title:         "7 Hari",
				Description:   "desc",
				Category:      categoryStreakMilestone,
				Threshold:     7,
				ProgressValue: 8,
				UnlockedAt:    now,
			},
		},
	}
	service := NewService(repo)
	service.now = func() time.Time { return now }

	payload, err := service.GetUnlocked(context.Background(), "user-1", CategoryQuery{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 unlocked item, got %d", len(payload.Items))
	}
	if payload.Items[0].AchievementCode != "streak_7_days" {
		t.Fatalf("unexpected code: %s", payload.Items[0].AchievementCode)
	}
}

func TestService_EnsureUserExists_NotFound(t *testing.T) {
	repo := &fakeAchievementsRepo{
		findUserErr: gorm.ErrRecordNotFound,
	}
	service := NewService(repo)

	_, err := service.GetCatalog(context.Background(), "missing-user", CategoryQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type fakeAchievementsRepo struct {
	user         models.User
	findUserErr  error
	achievements []models.Achievement
	metrics      evaluationMetrics
	metricsErr   error
	progressRows []progressListRow
	progressErr  error
	unlockedRows []unlockedListRow
	unlockedErr  error
	upsertRows   []progressUpsert
	upsertErr    error
}

func (r *fakeAchievementsRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.findUserErr != nil {
		return models.User{}, r.findUserErr
	}
	return r.user, nil
}

func (r *fakeAchievementsRepo) ListActiveAchievements(_ context.Context, _ *string) ([]models.Achievement, error) {
	return r.achievements, nil
}

func (r *fakeAchievementsRepo) ListProgressByUser(_ context.Context, _ string, _ *string) ([]progressListRow, error) {
	if r.progressErr != nil {
		return nil, r.progressErr
	}
	return r.progressRows, nil
}

func (r *fakeAchievementsRepo) ListUnlockedByUser(_ context.Context, _ string, _ *string) ([]unlockedListRow, error) {
	if r.unlockedErr != nil {
		return nil, r.unlockedErr
	}
	return r.unlockedRows, nil
}

func (r *fakeAchievementsRepo) ComputeEvaluationMetrics(_ context.Context, _ string, _ time.Time) (evaluationMetrics, error) {
	if r.metricsErr != nil {
		return evaluationMetrics{}, r.metricsErr
	}
	return r.metrics, nil
}

func (r *fakeAchievementsRepo) UpsertProgress(_ context.Context, _ string, rows []progressUpsert, _ time.Time) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upsertRows = append([]progressUpsert{}, rows...)
	return nil
}

var _ achievementsRepository = (*fakeAchievementsRepo)(nil)

func TestProgressForCategory_UnknownCategoryZero(t *testing.T) {
	got := progressForCategory("unknown", evaluationMetrics{StreakMilestone: 5})
	if got != 0 {
		t.Fatalf("expected 0 for unknown category, got %v", got)
	}
}

func TestIsRecordNotFound(t *testing.T) {
	if !IsRecordNotFound(gorm.ErrRecordNotFound) {
		t.Fatal("expected true for gorm.ErrRecordNotFound")
	}
	if IsRecordNotFound(errors.New("other")) {
		t.Fatal("expected false for non record-not-found error")
	}
}
