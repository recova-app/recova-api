package journals

import (
	"context"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

type journalsRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	CreateJournal(ctx context.Context, userID string, content string) (models.Journal, error)
	ListJournalsByUserID(ctx context.Context, userID string) ([]models.Journal, error)
}

// Service owns journals business rules.
type Service struct {
	repo journalsRepository
}

// NewService constructs journals service.
func NewService(repo journalsRepository) *Service {
	return &Service{repo: repo}
}

// CreateJournal stores one journal entry for owner user.
func (s *Service) CreateJournal(ctx context.Context, userID string, req CreateJournalRequest) (JournalPayload, error) {
	input, err := NormalizeCreateJournalRequest(req)
	if err != nil {
		return JournalPayload{}, err
	}

	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return JournalPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return JournalPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	row, err := s.repo.CreateJournal(ctx, userID, input.Content)
	if err != nil {
		return JournalPayload{}, errs.New(errs.CodeInternalError, "Gagal menyimpan jurnal", nil, err)
	}

	return mapJournalPayload(row), nil
}

// ListJournals returns user-scoped journal entries.
func (s *Service) ListJournals(ctx context.Context, userID string) ([]JournalPayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return nil, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListJournalsByUserID(ctx, userID)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca jurnal pengguna", nil, err)
	}

	payload := make([]JournalPayload, 0, len(rows))
	for _, row := range rows {
		payload = append(payload, mapJournalPayload(row))
	}
	return payload, nil
}

func mapJournalPayload(row models.Journal) JournalPayload {
	return JournalPayload{
		ID:        row.ID,
		UserID:    row.UserID,
		Content:   row.Content,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
	}
}
