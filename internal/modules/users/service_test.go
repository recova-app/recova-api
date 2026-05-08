package users

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
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
	checkIn := time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC)
	answers := map[string]any{"a": "b"}
	answersJSON, _ := json.Marshal(answers)

	repo := &fakeUsersRepo{
		user: models.User{
			ID:          "user-1",
			Email:       "user@example.test",
			Nickname:    "tester",
			UserWhy:     &reason,
			CheckInTime: &checkIn,
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
	checkIn := time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC)
	answersJSON, _ := json.Marshal(map[string]any{"a": "b"})

	repo := &fakeUsersRepo{
		user: models.User{
			ID:          "user-1",
			Email:       "user@example.test",
			Nickname:    "tester",
			UserWhy:     &reason,
			CheckInTime: &checkIn,
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

type fakeUsersRepo struct {
	user                  models.User
	profile               models.Profile
	profileErr            error
	updateErr             error
	completeOnboardingErr error
	resetErr              error
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

func (r *fakeUsersRepo) CompleteOnboarding(_ context.Context, _ string, _ OnboardingInput) (models.User, models.Profile, error) {
	if r.completeOnboardingErr != nil {
		return models.User{}, models.Profile{}, r.completeOnboardingErr
	}
	return r.user, models.Profile{ID: "profile-1", UserID: "user-1"}, nil
}

func (r *fakeUsersRepo) ResetUserDataForTesting(_ context.Context, _ string) error {
	return r.resetErr
}
