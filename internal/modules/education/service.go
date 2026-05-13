package education

import (
	"context"
	"errors"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

type educationRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	ListActiveContents(ctx context.Context) ([]models.EducationContent, error)
}

// Service owns education business rules.
type Service struct {
	repo educationRepository
}

// NewService constructs education service.
func NewService(repo educationRepository) *Service {
	return &Service{repo: repo}
}

// ListContents returns active education contents for authenticated user.
func (s *Service) ListContents(ctx context.Context, userID string) ([]ContentPayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListActiveContents(ctx)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca konten edukasi", nil, err)
	}

	payload := make([]ContentPayload, 0, len(rows))
	for _, row := range rows {
		payload = append(payload, ContentPayload{
			ID:           row.ID,
			Title:        row.Title,
			Description:  row.Description,
			URL:          row.URL,
			ThumbnailURL: row.ThumbnailURL,
			Category:     row.Category,
			PublishedAt:  formatPublishedAt(row.PublishedAt),
		})
	}

	return payload, nil
}
