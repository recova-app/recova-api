package users

import (
	"context"
	"encoding/json"
	"testing"

	aimodule "github.com/recova-app/backend-v2/internal/modules/ai"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

func TestService_UpdateSettings_ValidationErrorWhenEmptyPayload(t *testing.T) {
	svc := NewService(&fakeUsersRepo{}, "local", "development")

	_, err := svc.UpdateSettings(context.Background(), "user-1", SettingsUpdateRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_CompleteOnboarding_IdempotentWhenSamePayload(t *testing.T) {
	reason := "Fokus"
	check_in := "09:00"
	pornGoal := 7
	answers := map[string]any{"a": "b"}
	answersJSON, _ := json.Marshal(answers)

	repo := &fakeUsersRepo{
		user: models.User{
			ID:           "user-1",
			Email:        "user@example.test",
			Nickname:     "tester",
			UserWhy:      &reason,
			CheckInTime:  &check_in,
			PornFreeGoal: &pornGoal,
		},
		profile: models.Profile{
			ID:      "profile-1",
			UserID:  "user-1",
			Answers: answersJSON,
		},
	}
	svc := NewService(repo, "local", "development")

	payload, err := svc.CompleteOnboarding(context.Background(), "user-1", OnboardingRequest{
		Nickname:         "tester",
		RecoveryReason:   "Fokus",
		DailyCheckInTime: "09:00",
		PornFreeGoal:     intPointer(7),
		Answers:          answers,
	})
	if err != nil {
		t.Fatalf("complete onboarding: %v", err)
	}
	if !payload.OnboardingCompleted {
		t.Fatal("expected onboarding completed payload")
	}
}

func TestService_CompleteOnboarding_ConflictWhenAlreadyCompletedDifferentPayload(t *testing.T) {
	reason := "Fokus"
	check_in := "09:00"
	pornGoal := 7
	answersJSON, _ := json.Marshal(map[string]any{"a": "b"})

	repo := &fakeUsersRepo{
		user: models.User{
			ID:           "user-1",
			Email:        "user@example.test",
			Nickname:     "tester",
			UserWhy:      &reason,
			CheckInTime:  &check_in,
			PornFreeGoal: &pornGoal,
		},
		profile: models.Profile{
			ID:      "profile-1",
			UserID:  "user-1",
			Answers: answersJSON,
		},
	}
	svc := NewService(repo, "local", "development")

	_, err := svc.CompleteOnboarding(context.Background(), "user-1", OnboardingRequest{
		Nickname:         "tester-baru",
		RecoveryReason:   "Fokus",
		DailyCheckInTime: "09:00",
		PornFreeGoal:     intPointer(7),
		Answers:          map[string]any{"a": "b"},
	})
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestService_ResetUserDataForTesting_ForbiddenOutsideDev(t *testing.T) {
	svc := NewService(&fakeUsersRepo{}, "production", "production")

	err := svc.ResetUserDataForTesting(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestService_CompleteOnboarding_CallsAnalyzerAndReturnsSummary(t *testing.T) {
	repo := &fakeUsersRepo{
		user: models.User{
			ID:    "user-1",
			Email: "user@example.test",
		},
		profileErr: gorm.ErrRecordNotFound,
	}
	analyzer := &fakeOnboardingAnalyzer{
		response: aimodule.OnboardingAnalysisResponseData{
			Level:            "Moderate",
			Title:            "Perlu Konsistensi Bertahap",
			LevelDescription: "Konsistensi perlu diperkuat.",
			PatternAnalysis:  "Pemicu utama muncul saat stres.",
			Encouragement:    "Lanjutkan langkah kecil harian.",
		},
	}
	svc := NewService(repo, "local", "development", analyzer)

	payload, err := svc.CompleteOnboarding(context.Background(), "user-1", OnboardingRequest{
		Nickname:         "tester",
		RecoveryReason:   "Fokus",
		DailyCheckInTime: "09:00",
		PornFreeGoal:     intPointer(7),
		Answers: map[string]any{
			"motivation": "keluarga",
		},
	})
	if err != nil {
		t.Fatalf("complete onboarding: %v", err)
	}

	if analyzer.called != 1 {
		t.Fatalf("expected analyzer called once, got %d", analyzer.called)
	}
	if payload.OnboardingAnalysis == nil {
		t.Fatal("expected onboarding_analysis in payload")
	}
	if payload.OnboardingAnalysis.Level != "Moderate" {
		t.Fatalf("unexpected analysis level: %s", payload.OnboardingAnalysis.Level)
	}
	if repo.lastAISummary == nil || *repo.lastAISummary == "" {
		t.Fatal("expected ai summary persisted")
	}
	if repo.lastOnboardingInput.PornFreeGoal != 7 {
		t.Fatalf("unexpected porn_free_goal input: %d", repo.lastOnboardingInput.PornFreeGoal)
	}
}

func TestService_CompleteOnboarding_AnalyzerErrorBubblesUp(t *testing.T) {
	repo := &fakeUsersRepo{
		user:       models.User{ID: "user-1", Email: "user@example.test"},
		profileErr: gorm.ErrRecordNotFound,
	}
	analyzer := &fakeOnboardingAnalyzer{
		err: errs.New(errs.CodeServiceUnavailable, "AI unavailable", nil, nil),
	}
	svc := NewService(repo, "local", "development", analyzer)

	_, err := svc.CompleteOnboarding(context.Background(), "user-1", OnboardingRequest{
		Nickname:         "tester",
		RecoveryReason:   "Fokus",
		DailyCheckInTime: "09:00",
		PornFreeGoal:     intPointer(7),
		Answers:          map[string]any{"motivation": "keluarga"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeServiceUnavailable {
		t.Fatalf("unexpected error code: %s", mapped.Code)
	}
}

type fakeUsersRepo struct {
	user                  models.User
	profile               models.Profile
	profileErr            error
	updateErr             error
	completeOnboardingErr error
	resetErr              error
	lastOnboardingInput   OnboardingInput
	lastAISummary         *string
}

func (r *fakeUsersRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.user.ID == "" {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *fakeUsersRepo) FindProfileByUserID(_ context.Context, _ string) (models.Profile, error) {
	if r.profileErr != nil {
		return models.Profile{}, r.profileErr
	}
	if r.profile.ID == "" {
		return models.Profile{}, gorm.ErrRecordNotFound
	}
	return r.profile, nil
}

func (r *fakeUsersRepo) UpdateUserFields(_ context.Context, _ string, _ map[string]any) error {
	return r.updateErr
}

func (r *fakeUsersRepo) CompleteOnboarding(_ context.Context, _ string, input OnboardingInput, aiSummary *string) (models.User, models.Profile, error) {
	if r.completeOnboardingErr != nil {
		return models.User{}, models.Profile{}, r.completeOnboardingErr
	}
	r.lastOnboardingInput = input
	r.lastAISummary = aiSummary
	return r.user, models.Profile{ID: "profile-1", UserID: "user-1"}, nil
}

func (r *fakeUsersRepo) ResetUserDataForTesting(_ context.Context, _ string) error {
	return r.resetErr
}

type fakeOnboardingAnalyzer struct {
	response aimodule.OnboardingAnalysisResponseData
	err      error
	called   int
	lastReq  aimodule.OnboardingAnalysisRequest
}

func (a *fakeOnboardingAnalyzer) AnalyzeOnboarding(_ context.Context, _ string, req aimodule.OnboardingAnalysisRequest) (aimodule.OnboardingAnalysisResponseData, error) {
	a.called++
	a.lastReq = req
	if a.err != nil {
		return aimodule.OnboardingAnalysisResponseData{}, a.err
	}
	return a.response, nil
}

func intPointer(v int) *int {
	return &v
}
