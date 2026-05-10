package users

import "time"

// SettingsUpdateRequest contains mutable user settings payload.
type SettingsUpdateRequest struct {
	Nickname         *string `json:"nickname"`
	RecoveryReason   *string `json:"recovery_reason"`
	DailyCheckInTime *string `json:"daily_checkin_time"`
}

// OnboardingRequest contains onboarding payload under /auth/onboarding route.
type OnboardingRequest struct {
	Nickname         string         `json:"nickname"`
	RecoveryReason   string         `json:"recovery_reason"`
	DailyCheckInTime string         `json:"daily_checkin_time"`
	Answers          map[string]any `json:"answers"`
	DependencyLevel  string         `json:"dependency_level"`
}

// OnboardingInput is normalized onboarding payload for business layer.
type OnboardingInput struct {
	Nickname        string
	RecoveryReason  string
	DailyCheckInRaw string
	DailyCheckIn    time.Time
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
	OnboardingCompleted bool    `json:"onboarding_completed"`
}
