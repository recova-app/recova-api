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
		checkIns: []models.CheckIn{
			{CheckInDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC), IsSuccessful: true},
			{CheckInDate: time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC), IsSuccessful: true},
			{CheckInDate: time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC), IsSuccessful: true},
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

func TestService_GetStatistics_EnhancedFields(t *testing.T) {
	today := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	goal := 7
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester", PornFreeGoal: &goal},
		checkIns: []models.CheckIn{
			{CheckInDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC), Mood: "tenang", IsSuccessful: true},
			{CheckInDate: time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC), Mood: "fokus", IsSuccessful: true},
			{CheckInDate: time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC), Mood: "cemas", IsSuccessful: false},
			{CheckInDate: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Mood: "fokus", IsSuccessful: true},
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
	if payload.TotalAttempts != 4 {
		t.Fatalf("expected total attempts=4, got %d", payload.TotalAttempts)
	}
	if payload.SuccessRate != 0.75 {
		t.Fatalf("expected success rate=0.75, got %.2f", payload.SuccessRate)
	}
	if payload.RelapseCount != 1 {
		t.Fatalf("expected relapse count=1, got %d", payload.RelapseCount)
	}
	if payload.RelapseRate != 0.25 {
		t.Fatalf("expected relapse rate=0.25, got %.2f", payload.RelapseRate)
	}
	if payload.RecoverySuccessRate != 0.75 {
		t.Fatalf("expected recovery success rate=0.75, got %.2f", payload.RecoverySuccessRate)
	}
	if payload.CheckinConsistencyScore != 0.13 {
		t.Fatalf("expected consistency score=0.13, got %.2f", payload.CheckinConsistencyScore)
	}
	if len(payload.MoodTrend) != 4 {
		t.Fatalf("expected mood trend entries=4, got %d", len(payload.MoodTrend))
	}
	if payload.LastCheckInDayName == nil || *payload.LastCheckInDayName != "Jumat" {
		t.Fatalf("expected last check-in day name Jumat, got %+v", payload.LastCheckInDayName)
	}
	if payload.LastRelapseDayName == nil || *payload.LastRelapseDayName != "Kamis" {
		t.Fatalf("expected last relapse day name Kamis, got %+v", payload.LastRelapseDayName)
	}
	if len(payload.WeekdaySummary) != 7 {
		t.Fatalf("expected weekday summary size 7, got %d", len(payload.WeekdaySummary))
	}
	if payload.StreakGoalComparison.PornFreeGoal == nil || *payload.StreakGoalComparison.PornFreeGoal != 7 {
		t.Fatalf("expected porn_free_goal=7, got %+v", payload.StreakGoalComparison.PornFreeGoal)
	}
	if payload.StreakGoalComparison.RemainingDays == nil || *payload.StreakGoalComparison.RemainingDays != 6 {
		t.Fatalf("expected remaining days 6, got %+v", payload.StreakGoalComparison.RemainingDays)
	}
}

func TestService_GetActivitySummary_DefaultWindow(t *testing.T) {
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		windowRows: []models.CheckIn{
			{
				ID:           "checkin-1",
				CheckInDate:  time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
				Mood:         "tenang",
				IsSuccessful: true,
				CreatedAt:    time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC),
			},
			{
				ID:           "checkin-2",
				CheckInDate:  time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
				Mood:         "cemas",
				IsSuccessful: false,
				CreatedAt:    time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC),
			},
		},
		windowJournals: []models.Journal{
			{
				CheckInID: func() *string {
					v := "checkin-1"
					return &v
				}(),
				CreatedAt: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	svc := NewService(repo)
	svc.now = func() time.Time { return time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC) }

	payload, err := svc.GetActivitySummary(context.Background(), "user-1", ActivitySummaryQuery{})
	if err != nil {
		t.Fatalf("get activity summary: %v", err)
	}

	if payload.WindowDays != 30 {
		t.Fatalf("expected default window=30, got %d", payload.WindowDays)
	}
	if payload.SuccessfulCheckins != 1 {
		t.Fatalf("expected successful checkins=1, got %d", payload.SuccessfulCheckins)
	}
	if payload.Relapses != 1 {
		t.Fatalf("expected relapses=1, got %d", payload.Relapses)
	}
	if payload.ActiveDays != 2 {
		t.Fatalf("expected active days=2, got %d", payload.ActiveDays)
	}
	if len(payload.RecentActivity) != 3 {
		t.Fatalf("expected recent activity=3, got %d", len(payload.RecentActivity))
	}
	if payload.RecentActivity[0].Type != "journal" {
		t.Fatalf("expected most recent activity journal, got %s", payload.RecentActivity[0].Type)
	}
}

func TestService_GetRelapses_MapsJournalCommitment(t *testing.T) {
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		relapseRows: []models.CheckIn{
			{
				ID:             "checkin-1",
				CheckInDate:    time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
				Mood:           "cemas",
				IsSuccessful:   false,
				RelapseTrigger: []string{"stres kerja", "sendiri malam"},
				CreatedAt:      time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
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
	if len(payload[0].RelapseTrigger) != 2 || payload[0].RelapseTrigger[0] != "stres kerja" {
		t.Fatalf("expected relapse trigger mapped, got %+v", payload[0].RelapseTrigger)
	}
	if payload[0].CheckInDayName != "Jumat" {
		t.Fatalf("expected check-in day name Jumat, got %s", payload[0].CheckInDayName)
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
	checkIns         []models.CheckIn
	windowRows       []models.CheckIn
	windowJournals   []models.Journal
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
	listAllErr       error
	listWindowErr    error
	listJournalErr   error
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
func (r *fakeRoutineRepo) ListCheckInsByUser(_ context.Context, _ string) ([]models.CheckIn, error) {
	if r.listAllErr != nil {
		return nil, r.listAllErr
	}
	if len(r.checkIns) > 0 {
		return r.checkIns, nil
	}
	rows := make([]models.CheckIn, 0, len(r.successfulRows)+len(r.relapseRows))
	rows = append(rows, r.successfulRows...)
	rows = append(rows, r.relapseRows...)
	return rows, nil
}
func (r *fakeRoutineRepo) ListCheckInsByUserWithinDateRange(_ context.Context, _ string, _ time.Time, _ time.Time) ([]models.CheckIn, error) {
	if r.listWindowErr != nil {
		return nil, r.listWindowErr
	}
	return r.windowRows, nil
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
func (r *fakeRoutineRepo) ListJournalsByUserWithinTimeRange(_ context.Context, _ string, _ time.Time, _ time.Time) ([]models.Journal, error) {
	if r.listJournalErr != nil {
		return nil, r.listJournalErr
	}
	return r.windowJournals, nil
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
