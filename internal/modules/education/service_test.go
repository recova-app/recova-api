package education

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

type fakeEducationRepo struct {
	findUserErr   error
	listRows      []models.EducationContent
	listRowsError error
}

func (r *fakeEducationRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.findUserErr != nil {
		return models.User{}, r.findUserErr
	}
	return models.User{ID: "user-1"}, nil
}

func (r *fakeEducationRepo) ListActiveContents(_ context.Context) ([]models.EducationContent, error) {
	if r.listRowsError != nil {
		return nil, r.listRowsError
	}
	return r.listRows, nil
}

func TestService_ListContents_UserNotFound(t *testing.T) {
	svc := NewService(&fakeEducationRepo{findUserErr: gorm.ErrRecordNotFound})

	_, err := svc.ListContents(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mapped := errs.Map(err)
	if mapped.Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got: %s", mapped.Code)
	}
}

func TestService_ListContents_ListError(t *testing.T) {
	svc := NewService(&fakeEducationRepo{listRowsError: errors.New("db fail")})

	_, err := svc.ListContents(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mapped := errs.Map(err)
	if mapped.Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got: %s", mapped.Code)
	}
}

func TestService_ListContents_Success(t *testing.T) {
	published_at := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	svc := NewService(&fakeEducationRepo{
		listRows: []models.EducationContent{
			{
				ID:           "education-1",
				Title:        "judul",
				Description:  ptrString("deskripsi"),
				URL:          "https://example.test/edu",
				ThumbnailURL: ptrString("https://example.test/img.png"),
				Category:     "mindset",
				PublishedAt:  &published_at,
			},
		},
	})

	rows, err := svc.ListContents(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list contents: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row, got: %d", len(rows))
	}
	if rows[0].ID != "education-1" {
		t.Fatalf("unexpected id: %s", rows[0].ID)
	}
	if rows[0].PublishedAt == nil {
		t.Fatal("expected published_at payload")
	}
}

func ptrString(value string) *string {
	return &value
}
