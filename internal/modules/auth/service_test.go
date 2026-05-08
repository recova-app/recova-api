package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
)

func TestService_LoginWithGoogle_ValidationErrorWhenTokenEmpty(t *testing.T) {
	svc := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	_, err := svc.LoginWithGoogle(context.Background(), GoogleLoginRequest{Token: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_LoginWithGoogle_Success(t *testing.T) {
	repo := &fakeAuthRepo{
		userByGoogle: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
	}
	tokens := &fakeTokenProvider{
		accessPayload: SessionPayload{AccessToken: "access-1", TokenType: "Bearer", ExpiresIn: 900},
		refreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	verifier := &fakeGoogleVerifier{
		identity: GoogleIdentity{GoogleID: "google-1", Email: "user@example.test", DisplayName: "Tester"},
	}
	svc := NewService(repo, verifier, tokens)

	result, err := svc.LoginWithGoogle(context.Background(), GoogleLoginRequest{Token: "google-token"})
	if err != nil {
		t.Fatalf("login with google: %v", err)
	}
	if result.UserID != "user-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if repo.createdRefreshToken.TokenHash == "" {
		t.Fatal("expected refresh token persisted")
	}
}

func TestService_RefreshSession_Expired(t *testing.T) {
	repo := &fakeAuthRepo{
		storedRefresh: models.AuthRefreshToken{
			ID:        "token-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		},
	}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(-1*time.Minute))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for expired refresh")
	}
	if repo.revokedTokenID == "" {
		t.Fatal("expected stored token revoked")
	}
}

func TestService_BuildUserPayload(t *testing.T) {
	reason := "Fokus pemulihan"
	checkin := time.Date(2000, 1, 1, 9, 30, 0, 0, time.UTC)
	repo := &fakeAuthRepo{
		userByID: models.User{
			ID:          "user-1",
			Email:       "user@example.test",
			Nickname:    "tester",
			UserWhy:     &reason,
			CheckInTime: &checkin,
		},
		onboardingCompleted: true,
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	payload, err := svc.BuildUserPayload(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("build user payload: %v", err)
	}
	if !payload.OnboardingCompleted {
		t.Fatal("expected onboarding completed")
	}
	if payload.DailyCheckInTime == nil || *payload.DailyCheckInTime != "09:30" {
		t.Fatalf("unexpected check-in time: %#v", payload.DailyCheckInTime)
	}
}

type fakeGoogleVerifier struct {
	identity GoogleIdentity
	err      error
}

func (v *fakeGoogleVerifier) Verify(_ context.Context, _, _ string) (GoogleIdentity, error) {
	if v.err != nil {
		return GoogleIdentity{}, v.err
	}
	return v.identity, nil
}

type fakeTokenProvider struct {
	accessPayload      SessionPayload
	refreshClaims      SessionClaims
	parseAccessClaims  SessionClaims
	parseRefreshClaims SessionClaims
	parseAccessErr     error
	parseRefreshErr    error
}

func (f *fakeTokenProvider) GoogleAudience() string { return "google-client-id" }

func (f *fakeTokenProvider) IssueAccessToken(_ string) (string, SessionPayload, error) {
	if f.accessPayload.AccessToken == "" {
		f.accessPayload = SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}
	}
	return f.accessPayload.AccessToken, f.accessPayload, nil
}

func (f *fakeTokenProvider) IssueRefreshToken(_ string) (string, SessionClaims, error) {
	if f.refreshClaims.ExpiresAt == nil {
		f.refreshClaims = SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))}
	}
	return "refresh-token", f.refreshClaims, nil
}

func (f *fakeTokenProvider) ParseAccessToken(_ string) (SessionClaims, error) {
	if f.parseAccessErr != nil {
		return SessionClaims{}, f.parseAccessErr
	}
	if f.parseAccessClaims.Subject == "" {
		f.parseAccessClaims = SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "access-jti", time.Now().Add(time.Minute))}
	}
	return f.parseAccessClaims, nil
}

func (f *fakeTokenProvider) ParseRefreshToken(_ string) (SessionClaims, error) {
	if f.parseRefreshErr != nil {
		return SessionClaims{}, f.parseRefreshErr
	}
	if f.parseRefreshClaims.Subject == "" {
		f.parseRefreshClaims = SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))}
	}
	return f.parseRefreshClaims, nil
}

func (f *fakeTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }

func (f *fakeTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "recova_refresh", Value: rawToken}
}

func (f *fakeTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "recova_refresh", Value: "", MaxAge: -1}
}

func (f *fakeTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type fakeAuthRepo struct {
	userByGoogle        models.User
	userByID            models.User
	onboardingCompleted bool
	storedRefresh       models.AuthRefreshToken
	createdRefreshToken models.AuthRefreshToken
	revokedTokenID      string
	rotatedFromTokenID  string
}

func (r *fakeAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ GoogleIdentity) (models.User, error) {
	if r.userByGoogle.ID == "" {
		return models.User{}, errors.New("user missing")
	}
	return r.userByGoogle, nil
}

func (r *fakeAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.userByID.ID == "" {
		return models.User{}, errors.New("user missing")
	}
	return r.userByID, nil
}

func (r *fakeAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return r.onboardingCompleted, nil
}

func (r *fakeAuthRepo) CreateRefreshToken(_ context.Context, token models.AuthRefreshToken) error {
	r.createdRefreshToken = token
	return nil
}

func (r *fakeAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	if r.storedRefresh.ID == "" {
		return models.AuthRefreshToken{}, errors.New("refresh missing")
	}
	return r.storedRefresh, nil
}

func (r *fakeAuthRepo) RevokeRefreshTokenByID(_ context.Context, tokenID string, _ time.Time) error {
	r.revokedTokenID = tokenID
	return nil
}

func (r *fakeAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (r *fakeAuthRepo) RotateRefreshToken(_ context.Context, oldTokenID string, _ time.Time, _ models.AuthRefreshToken) error {
	r.rotatedFromTokenID = oldTokenID
	return nil
}

func jwtRegisteredClaims(subject string, id string, exp time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Subject:   subject,
		ID:        id,
		ExpiresAt: jwt.NewNumericDate(exp),
	}
}
