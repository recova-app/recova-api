package routine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

func TestNormalizeDailyCheckInRequest_ValidationError(t *testing.T) {
	_, err := NormalizeDailyCheckInRequest(DailyCheckInRequest{
		Mood: " ",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_GetStatistics_ComputesCurrentAndLongestStreak(t *testing.T) {
	today := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		successfulRows: []models.CheckIn{
			{CheckInDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)},
			{CheckInDate: time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)},
			{CheckInDate: time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC)},
		},
	}
	svc := NewService(repo)
	svc.now = func() time.Time { return today }

	payload, err := svc.GetStatistics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get statistics: %v", err)
	}

	if payload.TotalCheckins != 3 {
		t.Fatalf("expected total checkins=3, got %d", payload.TotalCheckins)
	}
	if payload.LongestStreak != 3 {
		t.Fatalf("expected longest streak=3, got %d", payload.LongestStreak)
	}
	if payload.CurrentStreak != 3 {
		t.Fatalf("expected current streak=3, got %d", payload.CurrentStreak)
	}
}

func TestService_GetRelapses_MapsJournalCommitment(t *testing.T) {
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		relapseRows: []models.CheckIn{
			{
				ID:           "checkin-1",
				CheckInDate:  time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
				Mood:         "cemas",
				IsSuccessful: false,
				CreatedAt:    time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
			},
		},
		journals: []models.Journal{
			{
				CheckInID: func() *string {
					v := "checkin-1"
					return &v
				}(),
				Content: "tetap tenang",
			},
		},
	}
	svc := NewService(repo)

	payload, err := svc.GetRelapses(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get relapses: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected one relapse row, got %d", len(payload))
	}
	if payload[0].Commitment == nil || *payload[0].Commitment != "tetap tenang" {
		t.Fatalf("expected commitment mapped, got %+v", payload[0].Commitment)
	}
}

func TestIsUniqueViolation_ReturnsTrueForCode23505(t *testing.T) {
	if !IsUniqueViolation(&pgconn.PgError{Code: uniqueViolationCode}) {
		t.Fatal("expected unique violation true for postgres code 23505")
	}
}

type fakeRoutineRepo struct {
	user             models.User
	findUserErr      error
	successfulRows   []models.CheckIn
	relapseRows      []models.CheckIn
	journals         []models.Journal
	latestBefore     *time.Time
	latestBeforeErr  error
	activeStreak     models.Streak
	activeStreakErr  error
	findCheckIn      models.CheckIn
	findCheckInErr   error
	createCheckInErr error
	createJournalErr error
	createStreakErr  error
	closeStreakErr   error
	listSuccessErr   error
	listRelapseErr   error
	findJournalsErr  error
}

func (r *fakeRoutineRepo) DB() *gorm.DB                         { return nil }
func (r *fakeRoutineRepo) CloneTx(_ *gorm.DB) routineRepository { return r }
func (r *fakeRoutineRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.findUserErr != nil {
		return models.User{}, r.findUserErr
	}
	if r.user.ID == "" {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return r.user, nil
}
func (r *fakeRoutineRepo) FindCheckInByUserAndDate(_ context.Context, _ string, _ time.Time) (models.CheckIn, error) {
	if r.findCheckInErr != nil {
		return models.CheckIn{}, r.findCheckInErr
	}
	return r.findCheckIn, nil
}
func (r *fakeRoutineRepo) CreateCheckIn(_ context.Context, _ models.CheckIn) error {
	return r.createCheckInErr
}
func (r *fakeRoutineRepo) CreateJournal(_ context.Context, _ models.Journal) error {
	return r.createJournalErr
}
func (r *fakeRoutineRepo) FindActiveStreak(_ context.Context, _ string) (models.Streak, error) {
	if r.activeStreakErr != nil {
		return models.Streak{}, r.activeStreakErr
	}
	if r.activeStreak.ID == "" {
		return models.Streak{}, gorm.ErrRecordNotFound
	}
	return r.activeStreak, nil
}
func (r *fakeRoutineRepo) CreateStreak(_ context.Context, _ models.Streak) error {
	return r.createStreakErr
}
func (r *fakeRoutineRepo) CloseActiveStreak(_ context.Context, _ string, _ time.Time) error {
	return r.closeStreakErr
}
func (r *fakeRoutineRepo) LatestSuccessfulCheckInBeforeDate(_ context.Context, _ string, _ time.Time) (*time.Time, error) {
	return r.latestBefore, r.latestBeforeErr
}
func (r *fakeRoutineRepo) ListSuccessfulCheckInsByUser(_ context.Context, _ string) ([]models.CheckIn, error) {
	if r.listSuccessErr != nil {
		return nil, r.listSuccessErr
	}
	return r.successfulRows, nil
}
func (r *fakeRoutineRepo) ListRelapseCheckInsByUser(_ context.Context, _ string) ([]models.CheckIn, error) {
	if r.listRelapseErr != nil {
		return nil, r.listRelapseErr
	}
	return r.relapseRows, nil
}
func (r *fakeRoutineRepo) FindJournalsByCheckInIDs(_ context.Context, _ []string) ([]models.Journal, error) {
	if r.findJournalsErr != nil {
		return nil, r.findJournalsErr
	}
	return r.journals, nil
}

func TestService_GetStatistics_NotFound(t *testing.T) {
	repo := &fakeRoutineRepo{findUserErr: gorm.ErrRecordNotFound}
	svc := NewService(repo)

	_, err := svc.GetStatistics(context.Background(), "missing-user")
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestService_GetRelapses_InternalError(t *testing.T) {
	repo := &fakeRoutineRepo{
		user:           models.User{ID: "user-1"},
		listRelapseErr: errors.New("db down"),
	}
	svc := NewService(repo)

	_, err := svc.GetRelapses(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected internal error")
	}
}
