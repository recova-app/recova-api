package journals

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

func TestNormalizeCreateJournalRequest_ValidationError(t *testing.T) {
	_, err := NormalizeCreateJournalRequest(CreateJournalRequest{Content: " "})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_CreateJournalSuccess(t *testing.T) {
	repo := &fakeJournalsRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		createRow: models.Journal{
			ID:        "journal-1",
			UserID:    "user-1",
			Content:   "catatan hari ini",
			CreatedAt: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
		},
	}
	svc := NewService(repo)

	payload, err := svc.CreateJournal(context.Background(), "user-1", CreateJournalRequest{
		Content: "catatan hari ini",
	})
	if err != nil {
		t.Fatalf("create journal: %v", err)
	}
	if payload.ID != "journal-1" {
		t.Fatalf("unexpected journal id: %s", payload.ID)
	}
}

func TestService_ListJournalsSuccess(t *testing.T) {
	repo := &fakeJournalsRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		listRows: []models.Journal{
			{ID: "journal-1", UserID: "user-1", Content: "satu", CreatedAt: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)},
			{ID: "journal-2", UserID: "user-1", Content: "dua", CreatedAt: time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)},
		},
	}
	svc := NewService(repo)

	rows, err := svc.ListJournals(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list journals: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 journals, got %d", len(rows))
	}
}

func TestService_ListJournalsInternalError(t *testing.T) {
	repo := &fakeJournalsRepo{
		user:    models.User{ID: "user-1"},
		listErr: errors.New("db down"),
	}
	svc := NewService(repo)

	_, err := svc.ListJournals(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected internal error")
	}
}

type fakeJournalsRepo struct {
	user      models.User
	findErr   error
	createRow models.Journal
	createErr error
	listRows  []models.Journal
	listErr   error
}

func (r *fakeJournalsRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.findErr != nil {
		return models.User{}, r.findErr
	}
	if r.user.ID == "" {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *fakeJournalsRepo) CreateJournal(_ context.Context, _ string, _ string) (models.Journal, error) {
	if r.createErr != nil {
		return models.Journal{}, r.createErr
	}
	return r.createRow, nil
}

func (r *fakeJournalsRepo) ListJournalsByUserID(_ context.Context, _ string) ([]models.Journal, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listRows, nil
}
