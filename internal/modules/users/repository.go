package users

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

const uniqueViolationCode = "23505"

// Repository provides persistence operations for users and onboarding.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs users repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTx clones repository to reuse transaction-scoped GORM handle.
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// DB returns underlying GORM handle.
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// FindUserByID loads user by id.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindProfileByUserID loads profile by user id.
func (r *Repository) FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error) {
	var profile models.Profile
	err := r.db.WithContext(ctx).Where("user_id = ?", strings.TrimSpace(userID)).First(&profile).Error
	if err != nil {
		return models.Profile{}, err
	}
	return profile, nil
}

// UpdateUserFields updates mutable user columns.
func (r *Repository) UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", strings.TrimSpace(userID)).
		Updates(fields).Error
}

// CreateProfile inserts onboarding profile row.
func (r *Repository) CreateProfile(ctx context.Context, profile models.Profile) error {
	return r.db.WithContext(ctx).Create(&profile).Error
}

// CompleteOnboarding creates profile and updates user in one transaction.
func (r *Repository) CompleteOnboarding(ctx context.Context, userID string, input OnboardingInput, aiSummary *string) (models.User, models.Profile, error) {
	var user models.User
	var profile models.Profile

	err := database.WithTransaction(ctx, r.db, func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)

		currentUser, err := txRepo.FindUserByID(ctx, userID)
		if err != nil {
			return err
		}
		user = currentUser

		if err := txRepo.UpdateUserFields(ctx, userID, map[string]any{
			"nickname":       input.Nickname,
			"user_why":       input.RecoveryReason,
			"check_in_time":  input.DailyCheckInRaw,
			"porn_free_goal": input.PornFreeGoal,
		}); err != nil {
			return err
		}

		answersJSON, err := json.Marshal(input.Answers)
		if err != nil {
			return err
		}

		newProfile := models.Profile{
			UserID:          strings.TrimSpace(userID),
			Answers:         answersJSON,
			DependencyLevel: input.DependencyLevel,
			AISummary:       aiSummary,
		}
		if err := txRepo.CreateProfile(ctx, newProfile); err != nil {
			return err
		}

		storedProfile, err := txRepo.FindProfileByUserID(ctx, userID)
		if err != nil {
			return err
		}
		profile = storedProfile

		freshUser, err := txRepo.FindUserByID(ctx, userID)
		if err != nil {
			return err
		}
		user = freshUser
		return nil
	})
	if err != nil {
		return models.User{}, models.Profile{}, err
	}

	return user, profile, nil
}

// ResetUserDataForTesting removes user-generated records and clears onboarding/profile state.
func (r *Repository) ResetUserDataForTesting(ctx context.Context, userID string) error {
	return database.WithTransaction(ctx, r.db, func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)
		trimmedUserID := strings.TrimSpace(userID)

		var userPosts []models.CommunityPost
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Find(&userPosts).Error; err != nil {
			return err
		}

		if len(userPosts) > 0 {
			postIDs := make([]string, 0, len(userPosts))
			for _, post := range userPosts {
				postIDs = append(postIDs, post.ID)
			}
			if err := tx.WithContext(ctx).Where("post_id IN ?", postIDs).Delete(&models.CommunityComment{}).Error; err != nil {
				return err
			}
		}

		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.CommunityComment{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.CommunityPostLike{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.CommunityPost{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.AIChat{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.Journal{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.CheckIn{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.Streak{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.Profile{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("user_id = ?", trimmedUserID).Delete(&models.AuthRefreshToken{}).Error; err != nil {
			return err
		}
		if err := txRepo.UpdateUserFields(ctx, trimmedUserID, map[string]any{
			"user_why":       nil,
			"check_in_time":  nil,
			"porn_free_goal": nil,
		}); err != nil {
			return err
		}
		return nil
	})
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueViolation reports postgres unique violation errors.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolationCode
}
