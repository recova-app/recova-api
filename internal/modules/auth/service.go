package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

type authRepository interface {
	FindOrCreateUserByGoogleIdentity(ctx context.Context, identity GoogleIdentity) (models.User, error)
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	IsOnboardingCompleted(ctx context.Context, userID string) (bool, error)
	CreateRefreshToken(ctx context.Context, token models.AuthRefreshToken) error
	GetActiveRefreshTokenByHash(ctx context.Context, tokenHash string) (models.AuthRefreshToken, error)
	RevokeRefreshTokenByID(ctx context.Context, tokenID string, revokedAt time.Time) error
	RevokeRefreshTokenByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
	RotateRefreshToken(ctx context.Context, oldTokenID string, revokedAt time.Time, newToken models.AuthRefreshToken) error
}

type tokenProvider interface {
	GoogleAudience() string
	IssueAccessToken(userID string) (string, SessionPayload, error)
	IssueRefreshToken(userID string) (string, SessionClaims, error)
	ParseAccessToken(rawToken string) (SessionClaims, error)
	ParseRefreshToken(rawToken string) (SessionClaims, error)
	HashRefreshToken(rawToken string) string
	RefreshCookie(rawToken string) *fiber.Cookie
	ExpiredRefreshCookie() *fiber.Cookie
	RefreshCookieValue(c fiber.Ctx) string
}

// Service owns authentication business rules and session orchestration.
type Service struct {
	repo     authRepository
	verifier GoogleTokenVerifier
	tokens   tokenProvider
}

// NewService constructs auth service with explicit dependencies.
func NewService(repo authRepository, verifier GoogleTokenVerifier, tokens tokenProvider) *Service {
	return &Service{
		repo:     repo,
		verifier: verifier,
		tokens:   tokens,
	}
}

// AuthenticateAccessToken validates bearer access token and returns request principal.
func (s *Service) AuthenticateAccessToken(rawToken string) (AuthPrincipal, error) {
	claims, err := s.tokens.ParseAccessToken(rawToken)
	if err != nil {
		return AuthPrincipal{}, errs.New(errs.CodeUnauthenticated, "Token akses tidak valid", nil, err)
	}

	return AuthPrincipal{UserID: strings.TrimSpace(claims.Subject)}, nil
}

// LoginWithGoogle verifies Google identity and issues new session tokens.
func (s *Service) LoginWithGoogle(ctx context.Context, req GoogleLoginRequest) (LoginResult, error) {
	if err := ValidateGoogleLoginRequest(req); err != nil {
		return LoginResult{}, err
	}

	identity, err := s.verifier.Verify(ctx, req.Token, s.tokens.GoogleAudience())
	if err != nil {
		return LoginResult{}, errs.New(errs.CodeUnauthenticated, "Token Google tidak valid", nil, err)
	}

	user, err := s.repo.FindOrCreateUserByGoogleIdentity(ctx, identity)
	if err != nil {
		if IsUniqueViolation(err) {
			return LoginResult{}, errs.New(errs.CodeConflict, "Akun pengguna mengalami konflik data", nil, err)
		}
		return LoginResult{}, errs.New(errs.CodeInternalError, "Gagal memproses login Google", nil, err)
	}

	accessToken, sessionPayload, err := s.tokens.IssueAccessToken(user.ID)
	if err != nil {
		return LoginResult{}, errs.New(errs.CodeInternalError, "Gagal membuat sesi akses", nil, err)
	}
	_ = accessToken

	refreshToken, refreshClaims, err := s.tokens.IssueRefreshToken(user.ID)
	if err != nil {
		return LoginResult{}, errs.New(errs.CodeInternalError, "Gagal membuat sesi refresh", nil, err)
	}

	if refreshClaims.ExpiresAt == nil {
		return LoginResult{}, errs.New(errs.CodeInternalError, "Token refresh tidak memiliki masa berlaku", nil, nil)
	}

	hash := s.tokens.HashRefreshToken(refreshToken)
	if err := s.repo.CreateRefreshToken(ctx, models.AuthRefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: refreshClaims.ExpiresAt.Time,
	}); err != nil {
		return LoginResult{}, errs.New(errs.CodeInternalError, "Gagal menyimpan sesi refresh", nil, err)
	}

	return LoginResult{
		UserID:           user.ID,
		RefreshToken:     refreshToken,
		RefreshTokenID:   refreshClaims.ID,
		RefreshExpiresAt: refreshClaims.ExpiresAt.Time,
		Session:          sessionPayload,
	}, nil
}

// RefreshSession validates refresh cookie, rotates refresh token, and issues fresh access token.
func (s *Service) RefreshSession(ctx context.Context, rawRefreshToken string) (RefreshSessionResult, error) {
	claims, err := s.tokens.ParseRefreshToken(rawRefreshToken)
	if err != nil {
		return RefreshSessionResult{}, errs.New(errs.CodeUnauthenticated, "Sesi login tidak valid", nil, err)
	}

	tokenHash := s.tokens.HashRefreshToken(rawRefreshToken)
	stored, err := s.repo.GetActiveRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if IsRecordNotFound(err) {
			return RefreshSessionResult{}, errs.New(errs.CodeUnauthenticated, "Sesi login tidak valid", nil, err)
		}
		return RefreshSessionResult{}, errs.New(errs.CodeInternalError, "Gagal membaca sesi refresh", nil, err)
	}

	if strings.TrimSpace(stored.UserID) != strings.TrimSpace(claims.Subject) {
		return RefreshSessionResult{}, errs.New(errs.CodeUnauthenticated, "Sesi login tidak valid", nil, fmt.Errorf("subject mismatch"))
	}

	now := time.Now().UTC()
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(now) || stored.ExpiresAt.Before(now) {
		_ = s.repo.RevokeRefreshTokenByID(ctx, stored.ID, now)
		return RefreshSessionResult{}, errs.New(errs.CodeUnauthenticated, "Sesi login sudah berakhir", nil, nil)
	}

	accessToken, sessionPayload, err := s.tokens.IssueAccessToken(stored.UserID)
	if err != nil {
		return RefreshSessionResult{}, errs.New(errs.CodeInternalError, "Gagal membuat sesi akses", nil, err)
	}
	_ = accessToken

	newRefreshToken, newClaims, err := s.tokens.IssueRefreshToken(stored.UserID)
	if err != nil {
		return RefreshSessionResult{}, errs.New(errs.CodeInternalError, "Gagal membuat sesi refresh", nil, err)
	}
	if newClaims.ExpiresAt == nil {
		return RefreshSessionResult{}, errs.New(errs.CodeInternalError, "Token refresh tidak memiliki masa berlaku", nil, nil)
	}

	rotatedFromID := stored.ID
	newRecord := models.AuthRefreshToken{
		UserID:        stored.UserID,
		TokenHash:     s.tokens.HashRefreshToken(newRefreshToken),
		ExpiresAt:     newClaims.ExpiresAt.Time,
		RotatedFromID: &rotatedFromID,
	}
	if err := s.repo.RotateRefreshToken(ctx, stored.ID, now, newRecord); err != nil {
		return RefreshSessionResult{}, errs.New(errs.CodeInternalError, "Gagal merotasi sesi refresh", nil, err)
	}

	return RefreshSessionResult{
		UserID:           stored.UserID,
		RefreshToken:     newRefreshToken,
		RefreshTokenID:   newClaims.ID,
		RefreshExpiresAt: newClaims.ExpiresAt.Time,
		Session:          sessionPayload,
	}, nil
}

// Logout revokes refresh token state when present and stays idempotent.
func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	trimmed := strings.TrimSpace(rawRefreshToken)
	if trimmed == "" {
		return nil
	}

	tokenHash := s.tokens.HashRefreshToken(trimmed)
	if err := s.repo.RevokeRefreshTokenByHash(ctx, tokenHash, time.Now().UTC()); err != nil {
		return errs.New(errs.CodeInternalError, "Gagal mengakhiri sesi login", nil, err)
	}

	return nil
}

// BuildAuthResponseData maps domain session state to API response payload.
func (s *Service) BuildAuthResponseData(ctx context.Context, userID string, session SessionPayload) (AuthResponseData, error) {
	userPayload, err := s.BuildUserPayload(ctx, userID)
	if err != nil {
		return AuthResponseData{}, err
	}

	return AuthResponseData{
		User:    userPayload,
		Session: session,
	}, nil
}

// BuildUserPayload loads current user profile summary for auth/users responses.
func (s *Service) BuildUserPayload(ctx context.Context, userID string) (UserPayload, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return UserPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return UserPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	completed, err := s.repo.IsOnboardingCompleted(ctx, userID)
	if err != nil {
		return UserPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca status onboarding", nil, err)
	}

	return UserPayload{
		ID:                  user.ID,
		Email:               user.Email,
		Nickname:            user.Nickname,
		RecoveryReason:      user.UserWhy,
		DailyCheckInTime:    formatCheckInTime(user.CheckInTime),
		OnboardingCompleted: completed,
	}, nil
}

// RefreshCookie builds secure refresh cookie for response header.
func (s *Service) RefreshCookie(rawToken string) *fiber.Cookie {
	return s.tokens.RefreshCookie(rawToken)
}

// ExpiredRefreshCookie builds cookie-expiration payload for logout response.
func (s *Service) ExpiredRefreshCookie() *fiber.Cookie {
	return s.tokens.ExpiredRefreshCookie()
}

// RefreshCookieValue extracts refresh token from request cookies.
func (s *Service) RefreshCookieValue(c fiber.Ctx) string {
	return s.tokens.RefreshCookieValue(c)
}

func formatCheckInTime(raw *time.Time) *string {
	if raw == nil {
		return nil
	}
	formatted := raw.UTC().Format(timeOfDayLayout)
	return &formatted
}
