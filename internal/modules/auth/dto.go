package auth

import "time"

// GoogleLoginRequest is request payload for Google login endpoint.
type GoogleLoginRequest struct {
	Token string `json:"token"`
}

// ManualRegisterRequest is request payload for manual register endpoint.
type ManualRegisterRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// ManualLoginRequest is request payload for manual login endpoint.
type ManualLoginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// SessionPayload is API payload for active access-token session.
type SessionPayload struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// UserPayload is API payload for authenticated user summary.
type UserPayload struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	Nickname            string  `json:"nickname"`
	RecoveryReason      *string `json:"recovery_reason"`
	DailyCheckInTime    *string `json:"daily_checkin_time"`
	PornFreeGoal        *int    `json:"porn_free_goal"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
}

// AuthResponseData contains combined user and session response payload.
type AuthResponseData struct {
	User    UserPayload    `json:"user"`
	Session SessionPayload `json:"session"`
}

// OnboardingRequest is request payload to complete onboarding.
type OnboardingRequest struct {
	Nickname         string         `json:"nickname"`
	RecoveryReason   string         `json:"recovery_reason"`
	DailyCheckInTime string         `json:"daily_checkin_time"`
	PornFreeGoal     *int           `json:"porn_free_goal"`
	Answers          map[string]any `json:"answers"`
	DependencyLevel  string         `json:"dependency_level"`
}

// OnboardingInput is normalized validated onboarding input for service layer.
type OnboardingInput struct {
	Nickname        string
	RecoveryReason  string
	DailyCheckInRaw string
	DailyCheckIn    time.Time
	PornFreeGoal    int
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

// ManualRegisterInput is normalized manual register request for service layer.
type ManualRegisterInput struct {
	Email    string
	Username string
	Password string
}

// ManualLoginInput is normalized manual login request for service layer.
type ManualLoginInput struct {
	Identifier string
	Password   string
}

// AuthPrincipal is request-scoped authenticated principal attached by middleware.
type AuthPrincipal struct {
	UserID string
}
