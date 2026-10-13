package journals

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

// Repository provides persistence operations for journals module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs journals repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByID loads user by identifier.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// CreateJournal inserts a personal journal entry.
func (r *Repository) CreateJournal(ctx context.Context, userID string, content string) (models.Journal, error) {
	row := models.Journal{
		UserID:  strings.TrimSpace(userID),
		Content: content,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return models.Journal{}, err
	}
	return row, nil
}

// ListJournalsByUserID returns journals by owner sorted latest first, bounded by limit and optional cursor.
func (r *Repository) ListJournalsByUserID(ctx context.Context, userID string, limit int, cursor *time.Time) ([]models.Journal, error) {
	query := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID))

	if cursor != nil {
		query = query.Where("created_at < ?", cursor.UTC())
	}

	if limit <= 0 || limit > maxJournalLimit {
		limit = defaultJournalLimit
	}

	var rows []models.Journal
	if err := query.
		Order("created_at desc").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
