package content

import (
	"context"
	"strings"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

// Repository provides persistence operations for daily content module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs content repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByID validates user existence by identifier.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// ListActiveMotivations returns active daily motivations sorted by stable order.
func (r *Repository) ListActiveMotivations(ctx context.Context) ([]models.DailyMotivation, error) {
	var rows []models.DailyMotivation
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at asc").
		Order("id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListActiveChallenges returns active daily challenges sorted by stable order.
func (r *Repository) ListActiveChallenges(ctx context.Context) ([]models.DailyChallenge, error) {
	var rows []models.DailyChallenge
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at asc").
		Order("id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
