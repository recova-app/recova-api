package education

import (
	"context"
	"strings"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

// Repository provides persistence operations for education module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs education repository.
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

// ListActiveContents returns active education content sorted by title.
func (r *Repository) ListActiveContents(ctx context.Context) ([]models.EducationContent, error) {
	var rows []models.EducationContent
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("title asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
