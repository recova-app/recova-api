package ai

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository provides persistence operations for AI module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs AI repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetCoachContext retrieves user, profile, active streak, chat history, and persona preference in consolidated queries.
func (r *Repository) GetCoachContext(ctx context.Context, userID string, historyLimit int) (CoachContext, error) {
	trimmedID := strings.TrimSpace(userID)

	type metaRow struct {
		UserID                 string
		Nickname               string
		UserWhy                *string
		ProfileID              *string
		ProfileAnswers         []byte
		ProfileDependencyLevel *string
		ProfileAISummary       *string
		StreakStartDate        *time.Time
		Persona                *string
	}

	var row metaRow
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(`
			u.id AS user_id,
			u.nickname,
			u.user_why,
			p.id AS profile_id,
			p.answers AS profile_answers,
			p.dependency_level AS profile_dependency_level,
			p.ai_summary AS profile_ai_summary,
			s.start_date AS streak_start_date,
			pref.persona
		`).
		Joins("LEFT JOIN profiles AS p ON p.user_id = u.id").
		Joins(`LEFT JOIN LATERAL (
			SELECT start_date
			FROM streaks
			WHERE user_id = u.id AND is_active = true
			ORDER BY start_date DESC
			LIMIT 1
		) AS s ON true`).
		Joins("LEFT JOIN user_ai_persona_preferences AS pref ON pref.user_id = u.id").
		Where("u.id = ?", trimmedID).
		Scan(&row).Error
	if err != nil {
		return CoachContext{}, err
	}
	if row.UserID == "" {
		return CoachContext{}, gorm.ErrRecordNotFound
	}

	chats, err := r.ListRecentChatsByUserID(ctx, trimmedID, historyLimit)
	if err != nil {
		return CoachContext{}, err
	}

	user := models.User{
		ID:       row.UserID,
		Nickname: row.Nickname,
		UserWhy:  row.UserWhy,
	}

	var profile models.Profile
	hasProfile := false
	if row.ProfileID != nil && *row.ProfileID != "" {
		hasProfile = true
		profile.ID = *row.ProfileID
		profile.UserID = row.UserID
		if row.ProfileAnswers != nil {
			profile.Answers = row.ProfileAnswers
		}
		profile.DependencyLevel = row.ProfileDependencyLevel
		profile.AISummary = row.ProfileAISummary
	}

	var activeStreak *models.Streak
	if row.StreakStartDate != nil {
		activeStreak = &models.Streak{
			UserID:    row.UserID,
			StartDate: *row.StreakStartDate,
			IsActive:  true,
		}
	}

	persona := DefaultPersona
	if row.Persona != nil && strings.TrimSpace(*row.Persona) != "" {
		persona = ResolvePersonaOrDefault(*row.Persona)
	}

	return CoachContext{
		User:         user,
		Profile:      profile,
		HasProfile:   hasProfile,
		ActiveStreak: activeStreak,
		RecentChats:  chats,
		Persona:      persona,
	}, nil
}

// FindUserByID loads user by id.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindProfileByUserID loads user profile by user id.
func (r *Repository) FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error) {
	var profile models.Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", strings.TrimSpace(userID)).First(&profile).Error; err != nil {
		return models.Profile{}, err
	}
	return profile, nil
}

// FindActiveStreakByUserID returns active streak row when available.
func (r *Repository) FindActiveStreakByUserID(ctx context.Context, userID string) (*models.Streak, error) {
	var row models.Streak
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("is_active = ?", true).
		Order("start_date desc").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListRecentChatsByUserID returns latest chat rows sorted ascending by created_at.
func (r *Repository) ListRecentChatsByUserID(ctx context.Context, userID string, limit int) ([]models.AIChat, error) {
	var rows []models.AIChat
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Order("created_at desc").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	reverseAIChats(rows)
	return rows, nil
}

// CreateChatMessages stores one or many AI chat rows.
func (r *Repository) CreateChatMessages(ctx context.Context, rows []models.AIChat) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

// GetPersonaPreferenceByUserID loads persona preference for one user.
func (r *Repository) GetPersonaPreferenceByUserID(ctx context.Context, userID string) (models.UserAIPersonaPreference, error) {
	var row models.UserAIPersonaPreference
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		First(&row).Error; err != nil {
		return models.UserAIPersonaPreference{}, err
	}
	return row, nil
}

// UpsertPersonaPreference stores persona preference idempotently by user id.
func (r *Repository) UpsertPersonaPreference(ctx context.Context, userID string, persona string, updatedAt time.Time) error {
	row := models.UserAIPersonaPreference{
		UserID:    strings.TrimSpace(userID),
		Persona:   strings.TrimSpace(persona),
		UpdatedAt: updatedAt.UTC(),
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"persona":    gorm.Expr("EXCLUDED.persona"),
				"updated_at": gorm.Expr("EXCLUDED.updated_at"),
			}),
		}).
		Create(&row).Error
}

func reverseAIChats(rows []models.AIChat) {
	for left, right := 0, len(rows)-1; left < right; left, right = left+1, right-1 {
		rows[left], rows[right] = rows[right], rows[left]
	}
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
