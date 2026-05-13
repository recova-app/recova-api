package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

const uniqueViolationCode = "23505"

// Repository provides persistence operations for auth/session state.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs auth repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindOrCreateUserByGoogleIdentity loads user by Google ID or creates a new one.
func (r *Repository) FindOrCreateUserByGoogleIdentity(ctx context.Context, identity GoogleIdentity) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("google_id = ?", strings.TrimSpace(identity.GoogleID)).First(&user).Error
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}

	nickname := strings.TrimSpace(identity.DisplayName)
	if nickname == "" {
		nickname = fallbackNickname(identity.Email)
	}
	googleID := strings.TrimSpace(identity.GoogleID)

	newUser := models.User{
		GoogleID: &googleID,
		Email:    strings.ToLower(strings.TrimSpace(identity.Email)),
		Nickname: nickname,
	}
	if err := r.db.WithContext(ctx).Create(&newUser).Error; err != nil {
		return models.User{}, err
	}

	return newUser, nil
}

// CreateManualUser inserts manual account row.
func (r *Repository) CreateManualUser(ctx context.Context, email string, username string, nickname string, passwordHash string) (models.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	if strings.TrimSpace(nickname) == "" {
		nickname = normalizedUsername
	}
	passwordHashTrimmed := strings.TrimSpace(passwordHash)

	newUser := models.User{
		Email:        normalizedEmail,
		Username:     &normalizedUsername,
		PasswordHash: &passwordHashTrimmed,
		Nickname:     strings.TrimSpace(nickname),
	}
	if err := r.db.WithContext(ctx).Create(&newUser).Error; err != nil {
		return models.User{}, err
	}

	return newUser, nil
}

// FindUserByLoginIdentifier finds one user by email or username.
func (r *Repository) FindUserByLoginIdentifier(ctx context.Context, identifier string) (models.User, error) {
	var user models.User
	normalized := strings.ToLower(strings.TrimSpace(identifier))
	if err := r.db.WithContext(ctx).
		Where("LOWER(email) = ?", normalized).
		Or("LOWER(username) = ?", normalized).
		First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindUserByID loads user by id.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// IsOnboardingCompleted returns true when profile row exists for user.
func (r *Repository) IsOnboardingCompleted(ctx context.Context, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Profile{}).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateRefreshToken inserts refresh token state row.
func (r *Repository) CreateRefreshToken(ctx context.Context, token models.AuthRefreshToken) error {
	return r.db.WithContext(ctx).Create(&token).Error
}

// GetActiveRefreshTokenByHash finds non-revoked refresh token by hash.
func (r *Repository) GetActiveRefreshTokenByHash(ctx context.Context, tokenHash string) (models.AuthRefreshToken, error) {
	var token models.AuthRefreshToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", strings.TrimSpace(tokenHash)).
		Where("revoked_at IS NULL").
		First(&token).Error
	if err != nil {
		return models.AuthRefreshToken{}, err
	}
	return token, nil
}

// RevokeRefreshTokenByID marks refresh token as revoked.
func (r *Repository) RevokeRefreshTokenByID(ctx context.Context, tokenID string, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.AuthRefreshToken{}).
		Where("id = ?", strings.TrimSpace(tokenID)).
		Where("revoked_at IS NULL").
		Updates(map[string]any{"revoked_at": revokedAt})
	return result.Error
}

// RevokeRefreshTokenByHash marks refresh token as revoked using its token hash.
func (r *Repository) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.AuthRefreshToken{}).
		Where("token_hash = ?", strings.TrimSpace(tokenHash)).
		Where("revoked_at IS NULL").
		Updates(map[string]any{"revoked_at": revokedAt})
	return result.Error
}

// RotateRefreshToken revokes old token and persists new token in one transaction.
func (r *Repository) RotateRefreshToken(ctx context.Context, oldTokenID string, revokedAt time.Time, newToken models.AuthRefreshToken) error {
	return database.WithTransaction(ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).
			Model(&models.AuthRefreshToken{}).
			Where("id = ?", strings.TrimSpace(oldTokenID)).
			Where("revoked_at IS NULL").
			Updates(map[string]any{"revoked_at": revokedAt}).
			Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Create(&newToken).Error; err != nil {
			return err
		}

		return nil
	})
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueViolation reports postgres unique-constraint violation.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolationCode
}

// UniqueViolationConstraint returns postgres constraint name when available.
func UniqueViolationConstraint(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ""
	}
	if pgErr.Code != uniqueViolationCode {
		return ""
	}
	return strings.TrimSpace(pgErr.ConstraintName)
}

func fallbackNickname(email string) string {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return "User"
	}

	parts := strings.Split(normalized, "@")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "User"
	}

	candidate := strings.TrimSpace(parts[0])
	if len([]rune(candidate)) < minNicknameLength {
		return "User"
	}
	if len([]rune(candidate)) > maxNicknameLength {
		runes := []rune(candidate)
		return string(runes[:maxNicknameLength])
	}
	return candidate
}
