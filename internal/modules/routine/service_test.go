package routine

import (
	"context"
	"errors"
	"strings"
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
	if len(payload.RelapseCalendar) != 1 || payload.RelapseCalendar[0] != "2026-05-07" {
		t.Fatalf("expected relapse calendar [2026-05-07], got %+v", payload.RelapseCalendar)
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

func TestService_GetStatistics_SameDayRelapseRemovesStreakContribution(t *testing.T) {
	today := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		checkIns: []models.CheckIn{
			{CheckInDate: time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC), Mood: "fokus", IsSuccessful: true},
			{CheckInDate: time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC), Mood: "tenang", IsSuccessful: true},
			{CheckInDate: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Mood: "cemas", IsSuccessful: true},
		},
		relapseRows: []models.Relapse{
			{RelapseDate: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Mood: "cemas"},
		},
	}
	svc := NewService(repo)
	svc.now = func() time.Time { return today }

	payload, err := svc.GetStatistics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get statistics: %v", err)
	}

	if payload.CurrentStreak != 0 {
		t.Fatalf("expected current streak=0 when relapse exists today, got %d", payload.CurrentStreak)
	}
	if payload.LongestStreak != 2 {
		t.Fatalf("expected longest streak=2 excluding relapse day, got %d", payload.LongestStreak)
	}
	if payload.TotalCheckins != 2 {
		t.Fatalf("expected total checkins=2 excluding relapse day success, got %d", payload.TotalCheckins)
	}
	if payload.RelapseCount != 1 {
		t.Fatalf("expected relapse count=1, got %d", payload.RelapseCount)
	}
	if len(payload.RelapseCalendar) != 1 || payload.RelapseCalendar[0] != "2026-05-08" {
		t.Fatalf("expected relapse calendar [2026-05-08], got %+v", payload.RelapseCalendar)
	}
	if payload.TotalAttempts != 3 {
		t.Fatalf("expected total attempts=3, got %d", payload.TotalAttempts)
	}
	if len(payload.StreakCalendar) != 2 || payload.StreakCalendar[0] != "2026-05-06" || payload.StreakCalendar[1] != "2026-05-07" {
		t.Fatalf("expected streak calendar without relapse day, got %+v", payload.StreakCalendar)
	}
}

func TestService_syncStreak_DoesNotOpenStreakWhenSameDayRelapseExists(t *testing.T) {
	repo := &fakeRoutineRepo{
		activeStreak: models.Streak{
			ID:        "streak-1",
			UserID:    "user-1",
			StartDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			IsActive:  true,
		},
		findRelapse: models.Relapse{
			ID:          "relapse-1",
			UserID:      "user-1",
			RelapseDate: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
		},
	}
	svc := NewService(repo)
	day := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)

	if err := svc.syncStreak(context.Background(), repo, "user-1", day, true); err != nil {
		t.Fatalf("sync streak: %v", err)
	}
	if len(repo.createdStreaks) != 0 {
		t.Fatalf("expected no streak creation on same-day relapse, got %d", len(repo.createdStreaks))
	}
	if len(repo.closedStreakIDs) != 1 || repo.closedStreakIDs[0] != "streak-1" {
		t.Fatalf("expected active streak closed once, got %+v", repo.closedStreakIDs)
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
				IsSuccessful: true,
				CreatedAt:    time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC),
			},
		},
		windowRelapses: []models.Relapse{
			{
				ID:          "relapse-1",
				RelapseDate: time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
				Mood:        "cemas",
				CreatedAt:   time.Date(2026, 5, 7, 11, 0, 0, 0, time.UTC),
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
	if payload.SuccessfulCheckins != 2 {
		t.Fatalf("expected successful checkins=2, got %d", payload.SuccessfulCheckins)
	}
	if payload.Relapses != 1 {
		t.Fatalf("expected relapses=1, got %d", payload.Relapses)
	}
	if payload.ActiveDays != 2 {
		t.Fatalf("expected active days=2, got %d", payload.ActiveDays)
	}
	if len(payload.RecentActivity) != 4 {
		t.Fatalf("expected recent activity=4, got %d", len(payload.RecentActivity))
	}
	if payload.RecentActivity[0].Type != "journal" {
		t.Fatalf("expected most recent activity journal, got %s", payload.RecentActivity[0].Type)
	}
}

func TestService_GetRelapses_MapsJournalCommitment(t *testing.T) {
	repo := &fakeRoutineRepo{
		user: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		relapseRows: []models.Relapse{
			{
				ID:             "relapse-1",
				UserID:         "user-1",
				RelapseDate:    time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
				Mood:           "cemas",
				Commitment:     func() *string { v := "tetap tenang"; return &v }(),
				RelapseTrigger: []string{"stres kerja", "sendiri malam"},
				CreatedAt:      time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
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
	if payload[0].RelapseDayName != "Jumat" {
		t.Fatalf("expected relapse day name Jumat, got %s", payload[0].RelapseDayName)
	}
}

func TestService_GetRelapseStatistics_BuildsHourlyPatternAndAISummary(t *testing.T) {
	aiSummary := "Kamu konsisten membaik, jaga ritme malam."
	repo := &fakeRoutineRepo{
		user:    models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		profile: models.Profile{ID: "profile-1", UserID: "user-1", AISummary: &aiSummary},
		checkIns: []models.CheckIn{
			{CheckInDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC), Mood: "tenang", IsSuccessful: true},
		},
		relapseRows: []models.Relapse{
			{
				ID:             "relapse-1",
				UserID:         "user-1",
				RelapseDate:    time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
				Mood:           "cemas",
				RelapseTrigger: []string{"sendiri malam"},
				CreatedAt:      time.Date(2026, 5, 11, 21, 5, 0, 0, time.UTC),
			},
			{
				ID:             "relapse-2",
				UserID:         "user-1",
				RelapseDate:    time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
				Mood:           "cemas",
				RelapseTrigger: []string{"stres kerja"},
				CreatedAt:      time.Date(2026, 5, 12, 21, 40, 0, 0, time.UTC),
			},
			{
				ID:             "relapse-3",
				UserID:         "user-1",
				RelapseDate:    time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC),
				Mood:           "gelisah",
				RelapseTrigger: []string{"bosan"},
				CreatedAt:      time.Date(2026, 5, 13, 22, 15, 0, 0, time.UTC),
			},
		},
	}
	svc := NewService(repo)
	svc.now = func() time.Time { return time.Date(2026, 5, 13, 23, 0, 0, 0, time.UTC) }

	payload, err := svc.GetRelapseStatistics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get relapse statistics: %v", err)
	}

	if payload.AISummary != aiSummary {
		t.Fatalf("expected ai summary mapped, got %q", payload.AISummary)
	}
	if payload.PeakRelapseCount != 2 {
		t.Fatalf("expected peak relapse count=2, got %d", payload.PeakRelapseCount)
	}
	if len(payload.PeakRelapseHoursUTC) != 1 || payload.PeakRelapseHoursUTC[0] != 21 {
		t.Fatalf("expected peak hour utc [21], got %+v", payload.PeakRelapseHoursUTC)
	}
	if len(payload.HourlyRelapseDistribution) != 2 {
		t.Fatalf("expected 2 hourly buckets, got %d", len(payload.HourlyRelapseDistribution))
	}
	if payload.HourlyRelapseDistribution[0].HourUTC != 21 || payload.HourlyRelapseDistribution[0].RelapseCount != 2 {
		t.Fatalf("expected hour 21 count 2, got %+v", payload.HourlyRelapseDistribution[0])
	}
	if payload.HourlyRelapseDistribution[1].HourUTC != 22 || payload.HourlyRelapseDistribution[1].RelapseCount != 1 {
		t.Fatalf("expected hour 22 count 1, got %+v", payload.HourlyRelapseDistribution[1])
	}
	if payload.RelapseTimeSummary.Title == "" || payload.RelapseTimeSummary.Analysis == "" || payload.RelapseTimeSummary.Summary == "" {
		t.Fatalf("expected relapse time summary with analysis+summary, got %+v", payload.RelapseTimeSummary)
	}
	if payload.RelapseTimeSummary.Title != "Analisis Waktu Relapse" {
		t.Fatalf("expected relapse time summary title updated, got %+v", payload.RelapseTimeSummary.Title)
	}
	if !strings.HasPrefix(payload.RelapseTimeSummary.Summary, "Trigger paling sering saat ini:") {
		t.Fatalf("expected summary mention top trigger, got %+v", payload.RelapseTimeSummary.Summary)
	}
	if payload.LatestRelapseSolution == nil {
		t.Fatal("expected latest relapse solution present")
	}
	if payload.LatestRelapseSolution.Summary == "" {
		t.Fatalf("expected latest relapse solution summary, got %+v", payload.LatestRelapseSolution)
	}
}

func TestService_GetRelapseStatistics_EmptyRelapseUsesFallbackSummary(t *testing.T) {
	repo := &fakeRoutineRepo{
		user:     models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"},
		checkIns: []models.CheckIn{},
	}
	svc := NewService(repo)
	svc.now = func() time.Time { return time.Date(2026, 5, 13, 23, 0, 0, 0, time.UTC) }

	payload, err := svc.GetRelapseStatistics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get relapse statistics: %v", err)
	}

	if payload.AISummary == "" {
		t.Fatal("expected default ai summary when profile missing")
	}
	if payload.PeakRelapseCount != 0 || len(payload.PeakRelapseHoursUTC) != 0 {
		t.Fatalf("expected no peak hours, got count=%d hours=%+v", payload.PeakRelapseCount, payload.PeakRelapseHoursUTC)
	}
	if payload.LatestRelapseSolution != nil {
		t.Fatalf("expected nil latest relapse solution, got %+v", payload.LatestRelapseSolution)
	}
	if payload.RelapseTimeSummary.Title == "" || payload.RelapseTimeSummary.Analysis == "" || payload.RelapseTimeSummary.Summary == "" {
		t.Fatalf("expected fallback relapse time summary analysis+summary, got %+v", payload.RelapseTimeSummary)
	}
}

func TestIsUniqueViolation_ReturnsTrueForCode23505(t *testing.T) {
	if !IsUniqueViolation(&pgconn.PgError{Code: uniqueViolationCode}) {
		t.Fatal("expected unique violation true for postgres code 23505")
	}
}

type fakeRoutineRepo struct {
	user              models.User
	findUserErr       error
	profile           models.Profile
	findProfileErr    error
	checkIns          []models.CheckIn
	windowRows        []models.CheckIn
	windowRelapses    []models.Relapse
	windowJournals    []models.Journal
	successfulRows    []models.CheckIn
	relapseRows       []models.Relapse
	journals          []models.Journal
	latestBefore      *time.Time
	latestBeforeErr   error
	activeStreak      models.Streak
	activeStreakErr   error
	findCheckIn       models.CheckIn
	findCheckInErr    error
	findRelapse       models.Relapse
	findRelapseErr    error
	createCheckInErr  error
	createRelapseErr  error
	createJournalErr  error
	upsertJournalErr  error
	createStreakErr   error
	closeStreakErr    error
	listAllErr        error
	listWindowErr     error
	listJournalErr    error
	listSuccessErr    error
	listRelapseErr    error
	listRelapseWndErr error
	createdStreaks    []models.Streak
	closedStreakIDs   []string
	closedStreakDates []time.Time
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
func (r *fakeRoutineRepo) FindProfileByUserID(_ context.Context, _ string) (models.Profile, error) {
	if r.findProfileErr != nil {
		return models.Profile{}, r.findProfileErr
	}
	if r.profile.ID == "" {
		return models.Profile{}, gorm.ErrRecordNotFound
	}
	return r.profile, nil
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
func (r *fakeRoutineRepo) UpsertJournalByCheckInID(_ context.Context, _ string, _ string, _ string) error {
	return r.upsertJournalErr
}
func (r *fakeRoutineRepo) FindRelapseByUserAndDate(_ context.Context, _ string, _ time.Time) (models.Relapse, error) {
	if r.findRelapseErr != nil {
		return models.Relapse{}, r.findRelapseErr
	}
	if r.findRelapse.ID == "" {
		return models.Relapse{}, gorm.ErrRecordNotFound
	}
	return r.findRelapse, nil
}
func (r *fakeRoutineRepo) CreateRelapse(_ context.Context, relapse models.Relapse) error {
	if r.createRelapseErr != nil {
		return r.createRelapseErr
	}
	r.findRelapse = relapse
	r.findRelapse.ID = "relapse-created"
	r.findRelapse.CreatedAt = time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	return nil
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
func (r *fakeRoutineRepo) CreateStreak(_ context.Context, streak models.Streak) error {
	if r.createStreakErr == nil {
		r.createdStreaks = append(r.createdStreaks, streak)
	}
	return r.createStreakErr
}
func (r *fakeRoutineRepo) CloseActiveStreak(_ context.Context, streakID string, endDate time.Time) error {
	if r.closeStreakErr == nil {
		r.closedStreakIDs = append(r.closedStreakIDs, streakID)
		r.closedStreakDates = append(r.closedStreakDates, endDate.UTC())
	}
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
	rows := make([]models.CheckIn, 0, len(r.successfulRows))
	rows = append(rows, r.successfulRows...)
	return rows, nil
}
func (r *fakeRoutineRepo) ListCheckInsByUserWithinDateRange(_ context.Context, _ string, _ time.Time, _ time.Time) ([]models.CheckIn, error) {
	if r.listWindowErr != nil {
		return nil, r.listWindowErr
	}
	return r.windowRows, nil
}
func (r *fakeRoutineRepo) ListRelapsesByUser(_ context.Context, _ string) ([]models.Relapse, error) {
	if r.listRelapseErr != nil {
		return nil, r.listRelapseErr
	}
	return r.relapseRows, nil
}
func (r *fakeRoutineRepo) ListRelapsesByUserWithinDateRange(_ context.Context, _ string, _ time.Time, _ time.Time) ([]models.Relapse, error) {
	if r.listRelapseWndErr != nil {
		return nil, r.listRelapseWndErr
	}
	return r.windowRelapses, nil
}
func (r *fakeRoutineRepo) FindJournalsByCheckInIDs(_ context.Context, _ []string) ([]models.Journal, error) {
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
