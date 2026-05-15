package users

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	aimodule "github.com/recova-app/backend-v2/internal/modules/ai"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

type usersRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error)
	UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error
	CompleteOnboarding(ctx context.Context, userID string, input OnboardingInput, aiSummary *string) (models.User, models.Profile, error)
	ResetUserDataForTesting(ctx context.Context, userID string) error
}

type onboardingAnalyzer interface {
	AnalyzeOnboarding(ctx context.Context, userID string, req aimodule.OnboardingAnalysisRequest) (aimodule.OnboardingAnalysisResponseData, error)
}

// Service owns users and onboarding business rules.
type Service struct {
	repo        usersRepository
	analyzer    onboardingAnalyzer
	allowDevOps bool
}

// NewService constructs users service with runtime guard configuration.
func NewService(repo usersRepository, appEnv string, nodeEnv string, analyzers ...onboardingAnalyzer) *Service {
	allowReset := strings.EqualFold(strings.TrimSpace(appEnv), "local") ||
		strings.EqualFold(strings.TrimSpace(nodeEnv), "development")

	var analyzer onboardingAnalyzer
	if len(analyzers) > 0 {
		analyzer = analyzers[0]
	}

	return &Service{
		repo:        repo,
		analyzer:    analyzer,
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
func (s *Service) CompleteOnboarding(ctx context.Context, userID string, req OnboardingRequest) (OnboardingCompletionPayload, error) {
	input, err := NormalizeOnboardingRequest(req)
	if err != nil {
		return OnboardingCompletionPayload{}, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return OnboardingCompletionPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return OnboardingCompletionPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	existingProfile, err := s.repo.FindProfileByUserID(ctx, userID)
	if err == nil {
		if sameOnboardingState(user, existingProfile, input) {
			return OnboardingCompletionPayload{
				UserProfilePayload: buildUserProfilePayload(user, true),
			}, nil
		}
		return OnboardingCompletionPayload{}, errs.New(errs.CodeConflict, "Pengguna sudah menyelesaikan onboarding", nil, nil)
	}
	if !IsRecordNotFound(err) {
		return OnboardingCompletionPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca status onboarding", nil, err)
	}

	if s.analyzer == nil {
		return OnboardingCompletionPayload{}, errs.New(errs.CodeInternalError, "Layanan analisis onboarding belum siap", nil, nil)
	}

	analysis, err := s.analyzer.AnalyzeOnboarding(ctx, userID, aimodule.OnboardingAnalysisRequest{
		Answers: input.Answers,
	})
	if err != nil {
		return OnboardingCompletionPayload{}, err
	}
	summary := composeOnboardingAISummary(analysis)

	updatedUser, _, err := s.repo.CompleteOnboarding(ctx, userID, input, &summary)
	if err != nil {
		if IsUniqueViolation(err) {
			return OnboardingCompletionPayload{}, errs.New(errs.CodeConflict, "Data onboarding mengalami konflik", nil, err)
		}
		return OnboardingCompletionPayload{}, errs.New(errs.CodeInternalError, "Gagal menyimpan onboarding", nil, err)
	}

	return OnboardingCompletionPayload{
		UserProfilePayload: buildUserProfilePayload(updatedUser, true),
		OnboardingAnalysis: &OnboardingAnalysisPayload{
			Level:            analysis.Level,
			Title:            analysis.Title,
			LevelDescription: analysis.LevelDescription,
			PatternAnalysis:  analysis.PatternAnalysis,
			Encouragement:    analysis.Encouragement,
		},
	}, nil
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
		PornFreeGoal:        copyIntPointer(user.PornFreeGoal),
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
	if user.PornFreeGoal == nil || *user.PornFreeGoal != input.PornFreeGoal {
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

func copyIntPointer(v *int) *int {
	if v == nil {
		return nil
	}
	copied := *v
	return &copied
}

func composeOnboardingAISummary(analysis aimodule.OnboardingAnalysisResponseData) string {
	segments := []string{
		strings.TrimSpace(analysis.Title),
		strings.TrimSpace(analysis.LevelDescription),
		strings.TrimSpace(analysis.PatternAnalysis),
		strings.TrimSpace(analysis.Encouragement),
	}

	filtered := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		filtered = append(filtered, segment)
	}
	if len(filtered) == 0 {
		return "Ringkasan onboarding belum tersedia."
	}

	return strings.TrimSpace(strings.Join(filtered, " "))
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
