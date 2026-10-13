package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
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

func TestService_RegisterManual_Success(t *testing.T) {
	repo := &fakeAuthRepo{
		manualCreatedUser: models.User{
			ID:       "user-manual-1",
			Email:    "manual@example.test",
			Nickname: "manualuser",
		},
	}
	tokens := &fakeTokenProvider{
		accessPayload: SessionPayload{AccessToken: "access-1", TokenType: "Bearer", ExpiresIn: 900},
		refreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-manual-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	result, err := svc.RegisterManual(context.Background(), ManualRegisterRequest{
		Email:           "manual@example.test",
		Username:        "manualuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err != nil {
		t.Fatalf("register manual: %v", err)
	}
	if result.UserID != "user-manual-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if repo.createdManualEmail != "manual@example.test" {
		t.Fatalf("unexpected created email: %s", repo.createdManualEmail)
	}
	if repo.createdManualPasswordHash == "" {
		t.Fatal("expected stored password hash")
	}
}

func TestService_RegisterManual_DuplicateUsername(t *testing.T) {
	repo := &fakeAuthRepo{
		createManualErr: &pgconn.PgError{
			Code:           uniqueViolationCode,
			ConstraintName: "uq_users_username",
		},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	_, err := svc.RegisterManual(context.Background(), ManualRegisterRequest{
		Email:           "manual@example.test",
		Username:        "manualuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeConflict {
		t.Fatalf("expected conflict app error, got: %v", err)
	}
}

func TestService_RegisterManual_DuplicateEmail(t *testing.T) {
	repo := &fakeAuthRepo{
		createManualErr: &pgconn.PgError{
			Code:           uniqueViolationCode,
			ConstraintName: "uq_users_email",
		},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	_, err := svc.RegisterManual(context.Background(), ManualRegisterRequest{
		Email:           "manual@example.test",
		Username:        "manualuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeConflict {
		t.Fatalf("expected conflict app error, got: %v", err)
	}
}

func TestService_LoginManual_NotFound(t *testing.T) {
	repo := &fakeAuthRepo{findManualErr: errors.New("not found")}
	tokens := &fakeTokenProvider{}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.LoginManual(context.Background(), ManualLoginRequest{
		Identifier: "manual@example.test",
		Password:   "password123",
	})
	if err == nil {
		t.Fatal("expected login error")
	}
}

func TestService_LoginManual_WrongPassword(t *testing.T) {
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repo := &fakeAuthRepo{
		manualUserByIdentifier: models.User{
			ID:           "user-manual-1",
			Email:        "manual@example.test",
			Nickname:     "manualuser",
			PasswordHash: &hash,
		},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	_, err = svc.LoginManual(context.Background(), ManualLoginRequest{
		Identifier: "manual@example.test",
		Password:   "wrong-password",
	})
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}
}

func TestService_LoginManual_Success(t *testing.T) {
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repo := &fakeAuthRepo{
		manualUserByIdentifier: models.User{
			ID:           "user-manual-1",
			Email:        "manual@example.test",
			Nickname:     "manualuser",
			PasswordHash: &hash,
		},
	}
	tokens := &fakeTokenProvider{
		accessPayload: SessionPayload{AccessToken: "access-1", TokenType: "Bearer", ExpiresIn: 900},
		refreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-manual-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	result, err := svc.LoginManual(context.Background(), ManualLoginRequest{
		Identifier: "manual@example.test",
		Password:   "password123",
	})
	if err != nil {
		t.Fatalf("login manual: %v", err)
	}
	if result.UserID != "user-manual-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
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

func TestService_RefreshSession_MalformedToken(t *testing.T) {
	tokens := &fakeTokenProvider{
		parseRefreshErr: errors.New("malformed jwt"),
	}
	svc := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "malformed-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for malformed token")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
}

func TestService_RefreshSession_TokenNotFound(t *testing.T) {
	repo := &fakeAuthRepo{getActiveRefreshErr: gorm.ErrRecordNotFound}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for missing token")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
}

func TestService_RefreshSession_SubjectMismatch(t *testing.T) {
	repo := &fakeAuthRepo{
		storedRefresh: models.AuthRefreshToken{
			ID:        "token-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("different-user", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for subject mismatch")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
}

func TestService_RefreshSession_JWTClaimsExpired(t *testing.T) {
	repo := &fakeAuthRepo{
		storedRefresh: models.AuthRefreshToken{
			ID:        "token-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(-1*time.Minute))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for expired jwt claims")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
	if repo.revokedTokenID != "token-1" {
		t.Fatalf("expected token-1 revoked on expiry, got: %s", repo.revokedTokenID)
	}
}

func TestService_RefreshSession_DatabaseTokenExpired(t *testing.T) {
	repo := &fakeAuthRepo{
		storedRefresh: models.AuthRefreshToken{
			ID:        "token-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		},
	}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for expired database token")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
	if repo.revokedTokenID != "token-1" {
		t.Fatalf("expected token-1 revoked on expiry, got: %s", repo.revokedTokenID)
	}
}

func TestService_RefreshSession_RotationAlreadyRevoked(t *testing.T) {
	repo := &fakeAuthRepo{
		storedRefresh: models.AuthRefreshToken{
			ID:        "token-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		rotateErr: ErrRefreshTokenAlreadyRevoked,
	}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected unauthenticated error for already revoked rotation")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
}

func TestService_RefreshSession_RotationUnexpectedError(t *testing.T) {
	repo := &fakeAuthRepo{
		storedRefresh: models.AuthRefreshToken{
			ID:        "token-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		rotateErr: errors.New("database connection down"),
	}
	tokens := &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	_, err := svc.RefreshSession(context.Background(), "refresh-token")
	if err == nil {
		t.Fatal("expected internal error for database failure")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeInternalError {
		t.Fatalf("expected internal error code, got: %v", err)
	}
}

func TestService_RefreshSession_ParallelConcurrentRequests(t *testing.T) {
	const totalRequests = 20
	initialToken := models.AuthRefreshToken{
		ID:        "token-1",
		UserID:    "user-1",
		TokenHash: "hash-refresh",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo := &concurrentRefreshRepo{
		token: initialToken,
	}
	tokens := &fakeTokenProvider{
		accessPayload: SessionPayload{AccessToken: "new-access", TokenType: "Bearer", ExpiresIn: 900},
		parseRefreshClaims: SessionClaims{
			RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour)),
		},
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, tokens)

	var wg sync.WaitGroup
	type callResult struct {
		result RefreshSessionResult
		err    error
	}
	results := make([]callResult, totalRequests)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			res, err := svc.RefreshSession(context.Background(), "refresh-token")
			results[idx] = callResult{result: res, err: err}
		}(i)
	}

	wg.Wait()

	successCount := 0
	unauthCount := 0
	for _, r := range results {
		if r.err == nil {
			successCount++
			if r.result.UserID != "user-1" {
				t.Errorf("unexpected user id: %s", r.result.UserID)
			}
		} else {
			var appErr *errs.AppError
			if errors.As(r.err, &appErr) && appErr.Code() == errs.CodeUnauthenticated {
				unauthCount++
			} else {
				t.Errorf("expected CodeUnauthenticated, got: %v", r.err)
			}
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful rotation, got %d", successCount)
	}
	if unauthCount != totalRequests-1 {
		t.Fatalf("expected %d unauthenticated failures, got %d", totalRequests-1, unauthCount)
	}
	if repo.rotatedCount != 1 {
		t.Fatalf("expected repo rotatedCount == 1, got %d", repo.rotatedCount)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected exactly 1 newly created token, got %d", len(repo.created))
	}
}

func TestService_AuthenticateAccessToken_Valid(t *testing.T) {
	tokens := &fakeTokenProvider{
		parseAccessClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-123", "access-jti", time.Now().Add(time.Minute))},
	}
	svc := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, tokens)

	principal, err := svc.AuthenticateAccessToken("valid-access-token")
	if err != nil {
		t.Fatalf("authenticate access token: %v", err)
	}
	if principal.UserID != "user-123" {
		t.Fatalf("unexpected user id: %s", principal.UserID)
	}
}

func TestService_AuthenticateAccessToken_Invalid(t *testing.T) {
	tokens := &fakeTokenProvider{
		parseAccessErr: errors.New("invalid signature"),
	}
	svc := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, tokens)

	_, err := svc.AuthenticateAccessToken("invalid-token")
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated error code, got: %v", err)
	}
}

func TestService_Logout_EmptyToken(t *testing.T) {
	repo := &fakeAuthRepo{}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	if err := svc.Logout(context.Background(), ""); err != nil {
		t.Fatalf("logout with empty token should succeed, got: %v", err)
	}
	if err := svc.Logout(context.Background(), "   "); err != nil {
		t.Fatalf("logout with whitespace token should succeed, got: %v", err)
	}
}

func TestService_Logout_ValidToken(t *testing.T) {
	repo := &fakeAuthRepo{}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	if err := svc.Logout(context.Background(), "valid-token"); err != nil {
		t.Fatalf("logout with valid token failed: %v", err)
	}
	if repo.revokedTokenHash != "hash-refresh" {
		t.Fatalf("expected revoked token hash %q, got %q", "hash-refresh", repo.revokedTokenHash)
	}
}

func TestService_Logout_RepositoryError(t *testing.T) {
	repo := &fakeAuthRepo{
		revokeByHashErr: errors.New("db error"),
	}
	svc := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	err := svc.Logout(context.Background(), "valid-token")
	if err == nil {
		t.Fatal("expected internal error")
	}
	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code() != errs.CodeInternalError {
		t.Fatalf("expected internal error code, got: %v", err)
	}
}

func TestService_BuildUserPayload(t *testing.T) {
	reason := "Recovery focus"
	checkin := "09:30"
	repo := &fakeAuthRepo{
		userByID: models.User{
			ID:          "user-1",
			Email:       "user@example.test",
			Nickname:    "tester",
			UserWhy:     &reason,
			CheckInTime: &checkin,
		},
		onboarding_completed: true,
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
	mu                 sync.RWMutex
	accessPayload      SessionPayload
	refreshClaims      SessionClaims
	parseAccessClaims  SessionClaims
	parseRefreshClaims SessionClaims
	parseAccessErr     error
	parseRefreshErr    error
	cookieValue        *string
}

func (f *fakeTokenProvider) GoogleAudience() string { return "google-client-id" }

func (f *fakeTokenProvider) IssueAccessToken(_ string) (string, SessionPayload, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.accessPayload.AccessToken == "" {
		f.accessPayload = SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}
	}
	return f.accessPayload.AccessToken, f.accessPayload, nil
}

func (f *fakeTokenProvider) IssueRefreshToken(_ string) (string, SessionClaims, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.refreshClaims.ExpiresAt == nil {
		f.refreshClaims = SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))}
	}
	return "refresh-token", f.refreshClaims, nil
}

func (f *fakeTokenProvider) ParseAccessToken(_ string) (SessionClaims, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.parseAccessErr != nil {
		return SessionClaims{}, f.parseAccessErr
	}
	if f.parseAccessClaims.Subject == "" {
		return SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "access-jti", time.Now().Add(time.Minute))}, nil
	}
	return f.parseAccessClaims, nil
}

func (f *fakeTokenProvider) ParseRefreshToken(_ string) (SessionClaims, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.parseRefreshErr != nil {
		return SessionClaims{}, f.parseRefreshErr
	}
	if f.parseRefreshClaims.Subject == "" {
		return SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(time.Hour))}, nil
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

func (f *fakeTokenProvider) RefreshCookieValue(c fiber.Ctx) string {
	if f.cookieValue != nil {
		return *f.cookieValue
	}
	if c != nil {
		val := strings.TrimSpace(c.Cookies("recova_refresh"))
		if val != "" {
			return val
		}
	}
	return "refresh-token"
}

type fakeAuthRepo struct {
	userByGoogle              models.User
	userByID                  models.User
	onboarding_completed      bool
	storedRefresh             models.AuthRefreshToken
	createdRefreshToken       models.AuthRefreshToken
	revokedTokenID            string
	rotatedFromTokenID        string
	manualCreatedUser         models.User
	manualUserByIdentifier    models.User
	createManualErr           error
	findManualErr             error
	rotateErr                 error
	revokeByHashErr           error
	getActiveRefreshErr       error
	revokedTokenHash          string
	createdManualEmail        string
	createdManualUsername     string
	createdManualNickname     string
	createdManualPasswordHash string
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
	return r.onboarding_completed, nil
}

func (r *fakeAuthRepo) CreateRefreshToken(_ context.Context, token models.AuthRefreshToken) error {
	r.createdRefreshToken = token
	return nil
}

func (r *fakeAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	if r.getActiveRefreshErr != nil {
		return models.AuthRefreshToken{}, r.getActiveRefreshErr
	}
	if r.storedRefresh.ID == "" {
		return models.AuthRefreshToken{}, gorm.ErrRecordNotFound
	}
	return r.storedRefresh, nil
}

func (r *fakeAuthRepo) RevokeRefreshTokenByID(_ context.Context, tokenID string, _ time.Time) error {
	r.revokedTokenID = tokenID
	return nil
}

func (r *fakeAuthRepo) RevokeRefreshTokenByHash(_ context.Context, tokenHash string, _ time.Time) error {
	if r.revokeByHashErr != nil {
		return r.revokeByHashErr
	}
	r.revokedTokenHash = tokenHash
	return nil
}

func (r *fakeAuthRepo) RotateRefreshToken(_ context.Context, oldTokenID string, _ time.Time, _ models.AuthRefreshToken) error {
	if r.rotateErr != nil {
		return r.rotateErr
	}
	r.rotatedFromTokenID = oldTokenID
	return nil
}

func (r *fakeAuthRepo) CreateManualUser(_ context.Context, email string, username string, nickname string, passwordHash string) (models.User, error) {
	if r.createManualErr != nil {
		return models.User{}, r.createManualErr
	}
	r.createdManualEmail = email
	r.createdManualUsername = username
	r.createdManualNickname = nickname
	r.createdManualPasswordHash = passwordHash
	if r.manualCreatedUser.ID == "" {
		return models.User{}, errors.New("manual user missing")
	}
	return r.manualCreatedUser, nil
}

func (r *fakeAuthRepo) FindUserByLoginIdentifier(_ context.Context, _ string) (models.User, error) {
	if r.findManualErr != nil {
		return models.User{}, r.findManualErr
	}
	if r.manualUserByIdentifier.ID == "" {
		return models.User{}, errors.New("manual user missing")
	}
	return r.manualUserByIdentifier, nil
}

func jwtRegisteredClaims(subject string, id string, exp time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Subject:   subject,
		ID:        id,
		ExpiresAt: jwt.NewNumericDate(exp),
	}
}

type concurrentRefreshRepo struct {
	mu           sync.Mutex
	token        models.AuthRefreshToken
	revoked      bool
	rotatedCount int
	created      []models.AuthRefreshToken
}

func (c *concurrentRefreshRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ GoogleIdentity) (models.User, error) {
	return models.User{}, errors.New("not implemented")
}

func (c *concurrentRefreshRepo) FindUserByID(_ context.Context, userID string) (models.User, error) {
	return models.User{ID: userID, Email: "user@example.test", Nickname: "user"}, nil
}

func (c *concurrentRefreshRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func (c *concurrentRefreshRepo) CreateRefreshToken(_ context.Context, token models.AuthRefreshToken) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.created = append(c.created, token)
	return nil
}

func (c *concurrentRefreshRepo) GetActiveRefreshTokenByHash(_ context.Context, hash string) (models.AuthRefreshToken, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.revoked || c.token.TokenHash != hash {
		return models.AuthRefreshToken{}, gorm.ErrRecordNotFound
	}
	return c.token, nil
}

func (c *concurrentRefreshRepo) RevokeRefreshTokenByID(_ context.Context, tokenID string, revokedAt time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token.ID == tokenID {
		c.revoked = true
		c.token.RevokedAt = &revokedAt
	}
	return nil
}

func (c *concurrentRefreshRepo) RevokeRefreshTokenByHash(_ context.Context, hash string, revokedAt time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token.TokenHash == hash {
		c.revoked = true
		c.token.RevokedAt = &revokedAt
	}
	return nil
}

func (c *concurrentRefreshRepo) RotateRefreshToken(_ context.Context, oldTokenID string, revokedAt time.Time, newToken models.AuthRefreshToken) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.revoked || c.token.ID != oldTokenID {
		return ErrRefreshTokenAlreadyRevoked
	}
	c.revoked = true
	c.token.RevokedAt = &revokedAt
	c.rotatedCount++
	c.created = append(c.created, newToken)
	return nil
}
