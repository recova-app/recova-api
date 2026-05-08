package auth

import "time"

// GoogleLoginRequest is request payload for Google login endpoint.
type GoogleLoginRequest struct {
	Token string `json:"token"`
}

// SessionPayload is API payload for active access-token session.
type SessionPayload struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// UserPayload is API payload for authenticated user summary.
type UserPayload struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	Nickname            string  `json:"nickname"`
	RecoveryReason      *string `json:"recoveryReason"`
	DailyCheckInTime    *string `json:"dailyCheckInTime"`
	OnboardingCompleted bool    `json:"onboardingCompleted"`
}

// AuthResponseData contains combined user and session response payload.
type AuthResponseData struct {
	User    UserPayload    `json:"user"`
	Session SessionPayload `json:"session"`
}

// OnboardingRequest is request payload to complete onboarding.
type OnboardingRequest struct {
	Nickname               string         `json:"nickname"`
	RecoveryReason         string         `json:"recovery_reason"`
	RecoveryReasonLegacy   string         `json:"userWhy"`
	DailyCheckInTime       string         `json:"daily_checkin_time"`
	DailyCheckInTimeLegacy string         `json:"checkinTime"`
	Answers                map[string]any `json:"answers"`
	DependencyLevel        string         `json:"dependency_level"`
	DependencyLevelLegacy  string         `json:"dependencyLevel"`
}

// OnboardingInput is normalized validated onboarding input for service layer.
type OnboardingInput struct {
	Nickname        string
	RecoveryReason  string
	DailyCheckInRaw string
	DailyCheckIn    time.Time
	Answers         map[string]any
	DependencyLevel *string
}

// GoogleIdentity contains safe claims extracted from verified Google token.
type GoogleIdentity struct {
	GoogleID    string
	Email       string
	DisplayName string
}

// RefreshSessionResult contains rotation result and token payload for response/cookie.
type RefreshSessionResult struct {
	UserID           string
	RefreshToken     string
	RefreshTokenID   string
	RefreshExpiresAt time.Time
	Session          SessionPayload
}

// LoginResult contains login output after Google verification.
type LoginResult struct {
	UserID           string
	RefreshToken     string
	RefreshTokenID   string
	RefreshExpiresAt time.Time
	Session          SessionPayload
}

// AuthPrincipal is request-scoped authenticated principal attached by middleware.
type AuthPrincipal struct {
	UserID string
}
