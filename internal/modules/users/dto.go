package users

import "time"

// SettingsUpdateRequest contains mutable user settings payload.
type SettingsUpdateRequest struct {
	Nickname         *string `json:"nickname"`
	RecoveryReason   *string `json:"recovery_reason"`
	DailyCheckInTime *string `json:"daily_checkin_time"`
	PornFreeGoal     *int    `json:"porn_free_goal"`
}

// OnboardingRequest contains onboarding payload under /auth/onboarding route.
type OnboardingRequest struct {
	Nickname         string         `json:"nickname"`
	RecoveryReason   string         `json:"recovery_reason"`
	DailyCheckInTime string         `json:"daily_checkin_time"`
	PornFreeGoal     *int           `json:"porn_free_goal"`
	Answers          map[string]any `json:"answers"`
	DependencyLevel  string         `json:"dependency_level"`
}

// OnboardingInput is normalized onboarding payload for business layer.
type OnboardingInput struct {
	Nickname        string
	RecoveryReason  string
	DailyCheckInRaw string
	DailyCheckIn    time.Time
	PornFreeGoal    int
	Answers         map[string]any
	DependencyLevel *string
}

// UserProfilePayload is users/onboarding API response payload.
type UserProfilePayload struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	Nickname            string  `json:"nickname"`
	RecoveryReason      *string `json:"recovery_reason"`
	DailyCheckInTime    *string `json:"daily_checkin_time"`
	PornFreeGoal        *int    `json:"porn_free_goal"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
}

// OnboardingAnalysisPayload is AI onboarding analysis payload included in onboarding response.
type OnboardingAnalysisPayload struct {
	Level            string `json:"level"`
	Title            string `json:"title"`
	LevelDescription string `json:"level_description"`
	PatternAnalysis  string `json:"pattern_analysis"`
	Encouragement    string `json:"encouragement"`
}

// OnboardingCompletionPayload is onboarding response payload including user profile and AI analysis.
type OnboardingCompletionPayload struct {
	UserProfilePayload
	OnboardingAnalysis *OnboardingAnalysisPayload `json:"onboarding_analysis"`
}
