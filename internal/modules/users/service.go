package users

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

type usersRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error)
	UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error
	CompleteOnboarding(ctx context.Context, userID string, input OnboardingInput) (models.User, models.Profile, error)
	ResetUserDataForTesting(ctx context.Context, userID string) error
}

// Service owns users and onboarding business rules.
type Service struct {
	repo        usersRepository
	allowDevOps bool
}

// NewService constructs users service with runtime guard configuration.
func NewService(repo usersRepository, appEnv string, nodeEnv string) *Service {
	allowReset := strings.EqualFold(strings.TrimSpace(appEnv), "local") ||
		strings.EqualFold(strings.TrimSpace(nodeEnv), "development")

	return &Service{
		repo:        repo,
		allowDevOps: allowReset,
	}
}

// GetCurrentUser returns current-user profile summary.
func (s *Service) GetCurrentUser(ctx context.Context, userID string) (UserProfilePayload, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return UserProfilePayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return UserProfilePayload{}, errs.New(errs.CodeInternalError, "Gagal membaca profil pengguna", nil, err)
	}

	completed, err := s.onboarding_completed(ctx, userID)
	if err != nil {
		return UserProfilePayload{}, err
	}

	return buildUserProfilePayload(user, completed), nil
}

// UpdateSettings updates mutable user settings and returns fresh user payload.
func (s *Service) UpdateSettings(ctx context.Context, userID string, req SettingsUpdateRequest) (UserProfilePayload, error) {
	updates, err := NormalizeSettingsUpdate(req)
	if err != nil {
		return UserProfilePayload{}, err
	}

	if err := s.repo.UpdateUserFields(ctx, userID, updates); err != nil {
		if IsRecordNotFound(err) {
			return UserProfilePayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		if IsUniqueViolation(err) {
			return UserProfilePayload{}, errs.New(errs.CodeConflict, "Data pengguna mengalami konflik", nil, err)
		}
		return UserProfilePayload{}, errs.New(errs.CodeInternalError, "Gagal memperbarui profil pengguna", nil, err)
	}

	return s.GetCurrentUser(ctx, userID)
}

// CompleteOnboarding validates onboarding payload and persists completion state.
func (s *Service) CompleteOnboarding(ctx context.Context, userID string, req OnboardingRequest) (UserProfilePayload, error) {
	input, err := NormalizeOnboardingRequest(req)
	if err != nil {
		return UserProfilePayload{}, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return UserProfilePayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return UserProfilePayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	existingProfile, err := s.repo.FindProfileByUserID(ctx, userID)
	if err == nil {
		if sameOnboardingState(user, existingProfile, input) {
			return buildUserProfilePayload(user, true), nil
		}
		return UserProfilePayload{}, errs.New(errs.CodeConflict, "Pengguna sudah menyelesaikan onboarding", nil, nil)
	}
	if !IsRecordNotFound(err) {
		return UserProfilePayload{}, errs.New(errs.CodeInternalError, "Gagal membaca status onboarding", nil, err)
	}

	updatedUser, _, err := s.repo.CompleteOnboarding(ctx, userID, input)
	if err != nil {
		if IsUniqueViolation(err) {
			return UserProfilePayload{}, errs.New(errs.CodeConflict, "Data onboarding mengalami konflik", nil, err)
		}
		return UserProfilePayload{}, errs.New(errs.CodeInternalError, "Gagal menyimpan onboarding", nil, err)
	}

	return buildUserProfilePayload(updatedUser, true), nil
}

// ResetUserDataForTesting clears user-generated data on allowed development env only.
func (s *Service) ResetUserDataForTesting(ctx context.Context, userID string) error {
	if !s.allowDevOps {
		return errs.New(errs.CodeForbidden, "Endpoint reset data hanya tersedia di lingkungan development", nil, nil)
	}

	if err := s.repo.ResetUserDataForTesting(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return errs.New(errs.CodeInternalError, "Gagal mereset data pengguna", nil, err)
	}

	return nil
}

func (s *Service) onboarding_completed(ctx context.Context, userID string) (bool, error) {
	_, err := s.repo.FindProfileByUserID(ctx, userID)
	if err == nil {
		return true, nil
	}
	if IsRecordNotFound(err) {
		return false, nil
	}
	return false, errs.New(errs.CodeInternalError, "Gagal membaca status onboarding", nil, err)
}

func buildUserProfilePayload(user models.User, completed bool) UserProfilePayload {
	return UserProfilePayload{
		ID:                  user.ID,
		Email:               user.Email,
		Nickname:            user.Nickname,
		RecoveryReason:      user.UserWhy,
		DailyCheckInTime:    formatCheckInTime(user.CheckInTime),
		OnboardingCompleted: completed,
	}
}

func formatCheckInTime(raw *string) *string {
	if raw == nil {
		return nil
	}
	formatted := normalizeTimeString(*raw)
	if formatted == "" {
		return nil
	}
	return &formatted
}

func sameOnboardingState(user models.User, profile models.Profile, input OnboardingInput) bool {
	if strings.TrimSpace(user.Nickname) != strings.TrimSpace(input.Nickname) {
		return false
	}
	if strings.TrimSpace(valueOrEmpty(user.UserWhy)) != strings.TrimSpace(input.RecoveryReason) {
		return false
	}
	if formatCheckInTime(user.CheckInTime) == nil {
		return false
	}
	if strings.TrimSpace(*formatCheckInTime(user.CheckInTime)) != strings.TrimSpace(input.DailyCheckInRaw) {
		return false
	}

	currentDependency := strings.TrimSpace(valueOrEmpty(profile.DependencyLevel))
	expectedDependency := strings.TrimSpace(valueOrEmpty(input.DependencyLevel))
	if currentDependency != expectedDependency {
		return false
	}

	var existingAnswers map[string]any
	if err := json.Unmarshal(profile.Answers, &existingAnswers); err != nil {
		return false
	}

	return reflect.DeepEqual(existingAnswers, input.Answers)
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func normalizeTimeString(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if len(trimmed) >= len(timeOfDayLayout) {
		return trimmed[:len(timeOfDayLayout)]
	}

	return trimmed
}
