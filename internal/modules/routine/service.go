package routine

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	aimodule "github.com/recova-app/backend-v2/internal/modules/ai"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

type routineRepository interface {
	DB() *gorm.DB
	CloneTx(tx *gorm.DB) routineRepository
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error)
	FindCheckInByUserAndDate(ctx context.Context, userID string, check_in_date time.Time) (models.CheckIn, error)
	CreateCheckIn(ctx context.Context, check_in models.CheckIn) error
	CreateJournal(ctx context.Context, journal models.Journal) error
	UpsertJournalByCheckInID(ctx context.Context, userID string, checkInID string, content string) error
	FindRelapseByUserAndDate(ctx context.Context, userID string, relapseDate time.Time) (models.Relapse, error)
	CreateRelapse(ctx context.Context, relapse models.Relapse) error
	ListRelapsesByUser(ctx context.Context, userID string) ([]models.Relapse, error)
	ListRelapsesByUserWithinDateRange(ctx context.Context, userID string, startDate time.Time, endDate time.Time) ([]models.Relapse, error)
	FindActiveStreak(ctx context.Context, userID string) (models.Streak, error)
	CreateStreak(ctx context.Context, streak models.Streak) error
	CloseActiveStreak(ctx context.Context, streakID string, endDate time.Time) error
	LatestSuccessfulCheckInBeforeDate(ctx context.Context, userID string, targetDate time.Time) (*time.Time, error)
	ListSuccessfulCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
	ListCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
	ListCheckInsByUserWithinDateRange(ctx context.Context, userID string, startDate time.Time, endDate time.Time) ([]models.CheckIn, error)
	FindJournalsByCheckInIDs(ctx context.Context, checkInIDs []string) ([]models.Journal, error)
	ListJournalsByUserWithinTimeRange(ctx context.Context, userID string, startAt time.Time, endAt time.Time) ([]models.Journal, error)
}

type relapseAdvisor interface {
	GenerateRelapseSolution(ctx context.Context, userID string, req aimodule.RelapseSolutionRequest) (aimodule.RelapseSolutionResponseData, error)
}

// Service owns routine business rules.
type Service struct {
	repo    routineRepository
	advisor relapseAdvisor
	now     func() time.Time
}

const defaultRoutineAISummaryMessage = "Insight baru untukmu akan segera tersedia. Teruslah menulis jurnal harianmu!"

// NewService constructs routine service with repository dependency.
func NewService(repo routineRepository, advisors ...relapseAdvisor) *Service {
	var advisor relapseAdvisor
	if len(advisors) > 0 {
		advisor = advisors[0]
	}

	return &Service{
		repo:    repo,
		advisor: advisor,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// CreateDailyCheckIn stores a daily check-in and updates streak state atomically.
func (s *Service) CreateDailyCheckIn(ctx context.Context, userID string, req DailyCheckInRequest) (CheckInResponseData, error) {
	input, err := NormalizeDailyCheckInRequest(req)
	if err != nil {
		return CheckInResponseData{}, err
	}

	check_in_date := utcDayStart(s.now())

	var stored models.CheckIn
	err = database.WithTransaction(ctx, s.repo.DB(), func(tx *gorm.DB) error {
		txRepo := s.repo.CloneTx(tx)

		if _, err := txRepo.FindUserByID(ctx, userID); err != nil {
			if IsRecordNotFound(err) {
				return errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
			}
			return errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
		}

		row := models.CheckIn{
			UserID:         strings.TrimSpace(userID),
			CheckInDate:    check_in_date,
			Mood:           input.Mood,
			IsSuccessful:   input.IsSuccessful,
			RelapseTrigger: nil,
		}
		if err := txRepo.CreateCheckIn(ctx, row); err != nil {
			if IsUniqueViolation(err) {
				return errs.New(errs.CodeConflict, "Pengguna sudah melakukan check-in hari ini", nil, err)
			}
			return errs.New(errs.CodeInternalError, "Gagal menyimpan check-in harian", nil, err)
		}

		created, err := txRepo.FindCheckInByUserAndDate(ctx, userID, check_in_date)
		if err != nil {
			return errs.New(errs.CodeInternalError, "Gagal membaca check-in terbaru", nil, err)
		}
		stored = created

		if input.JournalText != nil {
			if err := txRepo.CreateJournal(ctx, models.Journal{
				UserID:    strings.TrimSpace(userID),
				CheckInID: &stored.ID,
				Content:   *input.JournalText,
			}); err != nil {
				return errs.New(errs.CodeInternalError, "Gagal menyimpan catatan check-in", nil, err)
			}
		}

		if err := s.syncStreak(ctx, txRepo, userID, check_in_date, input.IsSuccessful); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return CheckInResponseData{}, err
	}

	stats, err := s.GetStatistics(ctx, userID)
	if err != nil {
		return CheckInResponseData{}, err
	}

	return CheckInResponseData{
		CheckIn:         mapCheckInPayload(stored, input.JournalText),
		Statistics:      stats,
		RelapseSolution: nil,
	}, nil
}

// CreateRelapse stores relapse trigger for current UTC day, independent from check-in submission route.
func (s *Service) CreateRelapse(ctx context.Context, userID string, req RelapseRequest) (RelapseResponseData, error) {
	input, err := NormalizeRelapseRequest(req)
	if err != nil {
		return RelapseResponseData{}, err
	}

	check_in_date := utcDayStart(s.now())

	var stored models.Relapse
	err = database.WithTransaction(ctx, s.repo.DB(), func(tx *gorm.DB) error {
		txRepo := s.repo.CloneTx(tx)

		if _, err := txRepo.FindUserByID(ctx, userID); err != nil {
			if IsRecordNotFound(err) {
				return errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
			}
			return errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
		}

		if _, err := txRepo.FindRelapseByUserAndDate(ctx, userID, check_in_date); err == nil {
			return errs.New(errs.CodeConflict, "Relapse hari ini sudah tercatat", nil, nil)
		} else if !IsRecordNotFound(err) {
			return errs.New(errs.CodeInternalError, "Gagal membaca relapse hari ini", nil, err)
		}

		var checkInID *string
		todayCheckIn, err := txRepo.FindCheckInByUserAndDate(ctx, userID, check_in_date)
		if err != nil && !IsRecordNotFound(err) {
			return errs.New(errs.CodeInternalError, "Gagal membaca check-in hari ini", nil, err)
		}
		if err == nil {
			checkInID = &todayCheckIn.ID
		}

		row := models.Relapse{
			UserID:         strings.TrimSpace(userID),
			CheckInID:      checkInID,
			RelapseDate:    check_in_date,
			Mood:           input.Mood,
			Commitment:     input.JournalText,
			RelapseTrigger: input.RelapseTrigger,
		}
		if err := txRepo.CreateRelapse(ctx, row); err != nil {
			if IsUniqueViolation(err) {
				return errs.New(errs.CodeConflict, "Relapse hari ini sudah tercatat", nil, err)
			}
			return errs.New(errs.CodeInternalError, "Gagal menyimpan relapse harian", nil, err)
		}

		if err := s.syncStreak(ctx, txRepo, userID, check_in_date, false); err != nil {
			return err
		}

		created, err := txRepo.FindRelapseByUserAndDate(ctx, userID, check_in_date)
		if err != nil {
			return errs.New(errs.CodeInternalError, "Gagal membaca relapse terbaru", nil, err)
		}
		stored = created
		return nil
	})
	if err != nil {
		return RelapseResponseData{}, err
	}

	stats, err := s.GetStatistics(ctx, userID)
	if err != nil {
		return RelapseResponseData{}, err
	}

	solution := s.buildRelapseSolution(ctx, userID, stored.Mood, stored.Commitment, stored.RelapseTrigger)

	return RelapseResponseData{
		Relapse:         mapRelapsePayload(stored),
		Statistics:      stats,
		RelapseSolution: &solution,
	}, nil
}

// GetStatistics returns routine statistics for authenticated user.
func (s *Service) GetStatistics(ctx context.Context, userID string) (StatisticsPayload, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return StatisticsPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return StatisticsPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListCheckInsByUser(ctx, userID)
	if err != nil {
		return StatisticsPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca statistik check-in", nil, err)
	}
	relapses, err := s.repo.ListRelapsesByUser(ctx, userID)
	if err != nil {
		return StatisticsPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca statistik relapse", nil, err)
	}

	return computeStatistics(rows, relapses, utcDayStart(s.now()), user.PornFreeGoal), nil
}

// GetActivitySummary returns periodic activity summary for authenticated user.
func (s *Service) GetActivitySummary(ctx context.Context, userID string, query ActivitySummaryQuery) (ActivitySummaryPayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return ActivitySummaryPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return ActivitySummaryPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	window_days, err := NormalizeActivitySummaryWindow(query.WindowDays)
	if err != nil {
		return ActivitySummaryPayload{}, err
	}

	endDate := utcDayStart(s.now())
	startDate := endDate.AddDate(0, 0, -(window_days - 1))
	checkIns, err := s.repo.ListCheckInsByUserWithinDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return ActivitySummaryPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca ringkasan aktivitas", nil, err)
	}
	relapses, err := s.repo.ListRelapsesByUserWithinDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return ActivitySummaryPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca ringkasan aktivitas", nil, err)
	}

	endExclusive := endDate.AddDate(0, 0, 1)
	journals, err := s.repo.ListJournalsByUserWithinTimeRange(ctx, userID, startDate, endExclusive)
	if err != nil {
		return ActivitySummaryPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca ringkasan aktivitas", nil, err)
	}

	return computeActivitySummary(window_days, checkIns, relapses, journals), nil
}

// GetRelapses returns user relapse history.
func (s *Service) GetRelapses(ctx context.Context, userID string) ([]RelapsePayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return nil, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListRelapsesByUser(ctx, userID)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca riwayat relapse", nil, err)
	}

	result := make([]RelapsePayload, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapRelapsePayload(row))
	}

	return result, nil
}

// GetRelapseStatistics returns complete relapse statistics with UTC hour pattern and AI suggestions.
func (s *Service) GetRelapseStatistics(ctx context.Context, userID string) (RelapseStatisticsResponseData, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return RelapseStatisticsResponseData{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return RelapseStatisticsResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	checkIns, err := s.repo.ListCheckInsByUser(ctx, userID)
	if err != nil {
		return RelapseStatisticsResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca statistik check-in", nil, err)
	}
	relapses, err := s.repo.ListRelapsesByUser(ctx, userID)
	if err != nil {
		return RelapseStatisticsResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca statistik relapse", nil, err)
	}

	stats := computeStatistics(checkIns, relapses, utcDayStart(s.now()), user.PornFreeGoal)
	relapseHistory := make([]RelapsePayload, 0, len(relapses))
	for _, row := range relapses {
		relapseHistory = append(relapseHistory, mapRelapsePayload(row))
	}

	hourlyDistribution, peakHoursUTC, peakCount := computeRelapseHourStats(relapses)
	triggerDistribution := computeRelapseTriggerStats(relapses)
	aiSummary, err := s.getCurrentAISummary(ctx, userID)
	if err != nil {
		return RelapseStatisticsResponseData{}, err
	}

	relapseTimeSummary := emptyRelapseTimeSummary(s.now())
	var relapseTriggerSummary *RelapseSolutionPayload
	if len(relapses) > 0 {
		latest := latestRelapseEvent(relapses)
		solution := s.buildRelapseSolution(ctx, userID, latest.Mood, latest.Commitment, latest.RelapseTrigger)
		relapseTriggerSummary = &solution
		relapseTimeSummary = s.buildRelapseTimeSummary(ctx, userID, relapses, peakHoursUTC, peakCount, latest)
	}

	return RelapseStatisticsResponseData{
		Statistics:                  stats,
		Relapses:                    relapseHistory,
		HourlyRelapseDistribution:   hourlyDistribution,
		RelapseTriggersDistribution: triggerDistribution,
		PeakRelapseHoursUTC:         peakHoursUTC,
		PeakRelapseCount:            peakCount,
		AISummary:                   aiSummary,
		RelapseTimeSummary:          relapseTimeSummary,
		RelapseTriggerSummary:       relapseTriggerSummary,
	}, nil
}

func (s *Service) syncStreak(ctx context.Context, repo routineRepository, userID string, check_in_date time.Time, is_successful bool) error {
	active, err := repo.FindActiveStreak(ctx, userID)
	if err != nil && !IsRecordNotFound(err) {
		return errs.New(errs.CodeInternalError, "Gagal membaca status streak", nil, err)
	}

	if !is_successful {
		if err == nil {
			if closeErr := repo.CloseActiveStreak(ctx, active.ID, check_in_date); closeErr != nil {
				return errs.New(errs.CodeInternalError, "Gagal menutup streak aktif", nil, closeErr)
			}
		}
		return nil
	}

	if _, relapseErr := repo.FindRelapseByUserAndDate(ctx, userID, check_in_date); relapseErr == nil {
		// Relapse on the same UTC day means this day cannot contribute to active streak.
		if active.ID != "" {
			if closeErr := repo.CloseActiveStreak(ctx, active.ID, check_in_date); closeErr != nil {
				return errs.New(errs.CodeInternalError, "Gagal menutup streak aktif", nil, closeErr)
			}
		}
		return nil
	} else if !IsRecordNotFound(relapseErr) {
		return errs.New(errs.CodeInternalError, "Gagal membaca data relapse harian", nil, relapseErr)
	}

	previousSuccessfulDate, err := repo.LatestSuccessfulCheckInBeforeDate(ctx, userID, check_in_date)
	if err != nil {
		return errs.New(errs.CodeInternalError, "Gagal membaca histori check-in", nil, err)
	}

	if previousSuccessfulDate == nil && active.ID == "" {
		if createErr := repo.CreateStreak(ctx, models.Streak{
			UserID:    strings.TrimSpace(userID),
			StartDate: check_in_date,
			IsActive:  true,
		}); createErr != nil {
			return errs.New(errs.CodeInternalError, "Gagal membuat streak baru", nil, createErr)
		}
		return nil
	}

	if active.ID == "" {
		if createErr := repo.CreateStreak(ctx, models.Streak{
			UserID:    strings.TrimSpace(userID),
			StartDate: check_in_date,
			IsActive:  true,
		}); createErr != nil {
			return errs.New(errs.CodeInternalError, "Gagal membuat streak baru", nil, createErr)
		}
		return nil
	}

	yesterday := check_in_date.AddDate(0, 0, -1)
	if previousSuccessfulDate != nil && sameUTCDate(*previousSuccessfulDate, yesterday) {
		return nil
	}

	if err := repo.CloseActiveStreak(ctx, active.ID, check_in_date); err != nil {
		return errs.New(errs.CodeInternalError, "Gagal menutup streak aktif", nil, err)
	}
	if err := repo.CreateStreak(ctx, models.Streak{
		UserID:    strings.TrimSpace(userID),
		StartDate: check_in_date,
		IsActive:  true,
	}); err != nil {
		return errs.New(errs.CodeInternalError, "Gagal membuat streak baru", nil, err)
	}

	return nil
}

func mapCheckInPayload(row models.CheckIn, commitment *string) CheckInPayload {
	return CheckInPayload{
		ID:             row.ID,
		UserID:         row.UserID,
		CheckInDate:    row.CheckInDate.UTC().Format("2006-01-02"),
		CheckInDayName: dayNameID(row.CheckInDate),
		Mood:           row.Mood,
		IsSuccessful:   row.IsSuccessful,
		Commitment:     commitment,
		RelapseTrigger: append([]string{}, row.RelapseTrigger...),
		CreatedAt:      row.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func mapRelapsePayload(row models.Relapse) RelapsePayload {
	return RelapsePayload{
		ID:             row.ID,
		UserID:         row.UserID,
		RelapseDate:    row.RelapseDate.UTC().Format("2006-01-02"),
		RelapseDayName: dayNameID(row.RelapseDate),
		Mood:           row.Mood,
		Commitment:     row.Commitment,
		RelapseTrigger: append([]string{}, row.RelapseTrigger...),
		CheckInID:      row.CheckInID,
		CreatedAt:      row.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func computeStatistics(checkIns []models.CheckIn, relapses []models.Relapse, todayUTC time.Time, pornFreeGoal *int) StatisticsPayload {
	if len(checkIns) == 0 && len(relapses) == 0 {
		return StatisticsPayload{
			CurrentStreak:           0,
			LongestStreak:           0,
			TotalCheckins:           0,
			TotalAttempts:           0,
			SuccessRate:             0,
			StreakCalendar:          []string{},
			RelapseCalendar:         []string{},
			RelapseCount:            0,
			RelapseRate:             0,
			RecoverySuccessRate:     0,
			CheckinConsistencyScore: 0,
			WeeklyProgress:          zeroProgressPayload(7),
			MonthlyProgress:         zeroProgressPayload(30),
			MoodTrend:               []MoodTrendPayload{},
			LastCheckInDate:         nil,
			LastCheckInDayName:      nil,
			LastRelapseDate:         nil,
			LastRelapseDayName:      nil,
			WeekdaySummary:          buildWeekdaySummary(nil, nil),
			StreakGoalComparison:    buildStreakGoalComparison(pornFreeGoal, 0, 0),
		}
	}

	successDates := make([]time.Time, 0, len(checkIns))
	calendarDays := make(map[string]struct{}, len(checkIns))
	active_days := map[string]struct{}{}
	relapseDays := map[string]struct{}{}
	legacyRelapseDays := map[string]struct{}{}
	moodByDay := map[string]map[string]int{}
	successCountByDay := map[string]int{}
	totalCountByDay := map[string]int{}
	var lastCheckInDate *time.Time
	var lastRelapseDate *time.Time

	for _, row := range checkIns {
		day := utcDayStart(row.CheckInDate)
		dayKey := day.Format("2006-01-02")
		active_days[dayKey] = struct{}{}
		totalCountByDay[dayKey]++
		lastCheckInDate = maxDayPtr(lastCheckInDate, day)

		if _, exists := moodByDay[dayKey]; !exists {
			moodByDay[dayKey] = map[string]int{}
		}
		moodKey := strings.TrimSpace(strings.ToLower(row.Mood))
		if moodKey == "" {
			moodKey = "unknown"
		}
		moodByDay[dayKey][moodKey]++

		if !row.IsSuccessful {
			legacyRelapseDays[dayKey] = struct{}{}
			relapseDays[dayKey] = struct{}{}
			lastRelapseDate = maxDayPtr(lastRelapseDate, day)
		}
	}

	for _, row := range relapses {
		day := utcDayStart(row.RelapseDate)
		dayKey := day.Format("2006-01-02")
		active_days[dayKey] = struct{}{}
		lastRelapseDate = maxDayPtr(lastRelapseDate, day)
		relapseDays[dayKey] = struct{}{}

		if _, fromLegacy := legacyRelapseDays[dayKey]; fromLegacy {
			continue
		}

		totalCountByDay[dayKey]++

		if _, exists := moodByDay[dayKey]; !exists {
			moodByDay[dayKey] = map[string]int{}
		}
		moodKey := strings.TrimSpace(strings.ToLower(row.Mood))
		if moodKey == "" {
			moodKey = "unknown"
		}
		moodByDay[dayKey][moodKey]++
	}

	successCount := 0
	for _, row := range checkIns {
		if !row.IsSuccessful {
			continue
		}
		day := utcDayStart(row.CheckInDate)
		dayKey := day.Format("2006-01-02")
		if _, blocked := relapseDays[dayKey]; blocked {
			continue
		}
		successCount++
		successCountByDay[dayKey]++
		successDates = append(successDates, day)
		calendarDays[dayKey] = struct{}{}
	}

	calendar := make([]string, 0, len(calendarDays))
	for day := range calendarDays {
		calendar = append(calendar, day)
	}
	sort.Strings(calendar)
	relapseCalendar := make([]string, 0, len(relapseDays))
	for day := range relapseDays {
		relapseCalendar = append(relapseCalendar, day)
	}
	sort.Strings(relapseCalendar)

	relapse_count := len(relapseDays)
	totalAttempts := successCount + relapse_count
	recovery_success_rate := safeRatio(successCount, totalAttempts)
	relapse_rate := safeRatio(relapse_count, totalAttempts)
	success_rate := safeRatio(successCount, totalAttempts)

	current_streak, longest_streak := computeStreaks(successDates, todayUTC)
	if _, hasRelapseToday := relapseDays[todayUTC.Format("2006-01-02")]; hasRelapseToday {
		current_streak = 0
	}
	checkin_consistency_score := safeRatio(len(active_days), 30)
	weekly_progress := computeProgressPayload(checkIns, todayUTC, 7)
	monthly_progress := computeProgressPayload(checkIns, todayUTC, 30)
	mood_trend := buildMoodTrendPayload(moodByDay, successCountByDay, totalCountByDay)
	weekdaySummary := buildWeekdaySummary(checkIns, relapses)
	lastCheckInDateRaw, lastCheckInDayName := mapLastDatePayload(lastCheckInDate)
	lastRelapseDateRaw, lastRelapseDayName := mapLastDatePayload(lastRelapseDate)

	return StatisticsPayload{
		CurrentStreak:           current_streak,
		LongestStreak:           longest_streak,
		TotalCheckins:           successCount,
		TotalAttempts:           totalAttempts,
		SuccessRate:             success_rate,
		StreakCalendar:          calendar,
		RelapseCalendar:         relapseCalendar,
		RelapseCount:            relapse_count,
		RelapseRate:             relapse_rate,
		RecoverySuccessRate:     recovery_success_rate,
		CheckinConsistencyScore: checkin_consistency_score,
		WeeklyProgress:          weekly_progress,
		MonthlyProgress:         monthly_progress,
		MoodTrend:               mood_trend,
		LastCheckInDate:         lastCheckInDateRaw,
		LastCheckInDayName:      lastCheckInDayName,
		LastRelapseDate:         lastRelapseDateRaw,
		LastRelapseDayName:      lastRelapseDayName,
		WeekdaySummary:          weekdaySummary,
		StreakGoalComparison:    buildStreakGoalComparison(pornFreeGoal, current_streak, longest_streak),
	}
}

func computeActivitySummary(window_days int, checkIns []models.CheckIn, relapses []models.Relapse, journals []models.Journal) ActivitySummaryPayload {
	successful_checkins := 0
	relapseCount := 0
	active_days := map[string]struct{}{}
	checkinByID := make(map[string]models.CheckIn, len(checkIns))
	activities := make([]activityTimelineItem, 0, len(checkIns)+len(journals))

	for _, row := range checkIns {
		day := utcDayStart(row.CheckInDate).Format("2006-01-02")
		active_days[day] = struct{}{}
		checkinByID[strings.TrimSpace(row.ID)] = row

		if row.IsSuccessful {
			successful_checkins++
		}
		if !row.IsSuccessful {
			relapseCount++
		}

		eventType := "checkin_success"
		if !row.IsSuccessful {
			eventType = "checkin_relapse"
		}
		mood := strings.TrimSpace(row.Mood)
		var moodPtr *string
		if mood != "" {
			value := mood
			moodPtr = &value
		}

		activities = append(activities, activityTimelineItem{
			Date:      day,
			DayName:   dayNameID(row.CheckInDate),
			Type:      eventType,
			Mood:      moodPtr,
			Timestamp: row.CreatedAt.UTC(),
		})
	}

	for _, relapse := range relapses {
		day := utcDayStart(relapse.RelapseDate).Format("2006-01-02")
		active_days[day] = struct{}{}
		relapseCount++

		mood := strings.TrimSpace(relapse.Mood)
		var moodPtr *string
		if mood != "" {
			value := mood
			moodPtr = &value
		}

		activities = append(activities, activityTimelineItem{
			Date:      day,
			DayName:   dayNameID(relapse.RelapseDate),
			Type:      "relapse",
			Mood:      moodPtr,
			Timestamp: relapse.CreatedAt.UTC(),
		})
	}

	for _, journal := range journals {
		day := utcDayStart(journal.CreatedAt).Format("2006-01-02")
		active_days[day] = struct{}{}

		var moodPtr *string
		if journal.CheckInID != nil {
			if row, exists := checkinByID[strings.TrimSpace(*journal.CheckInID)]; exists {
				mood := strings.TrimSpace(row.Mood)
				if mood != "" {
					value := mood
					moodPtr = &value
				}
			}
		}

		activities = append(activities, activityTimelineItem{
			Date:      day,
			DayName:   dayNameID(journal.CreatedAt),
			Type:      "journal",
			Mood:      moodPtr,
			Timestamp: journal.CreatedAt.UTC(),
		})
	}

	sort.SliceStable(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	recent_activity := make([]ActivityItemPayload, 0, len(activities))
	for _, item := range activities {
		recent_activity = append(recent_activity, ActivityItemPayload{
			Date:    item.Date,
			DayName: item.DayName,
			Type:    item.Type,
			Mood:    item.Mood,
		})
	}

	return ActivitySummaryPayload{
		WindowDays:         window_days,
		SuccessfulCheckins: successful_checkins,
		Relapses:           relapseCount,
		ActiveDays:         len(active_days),
		RecentActivity:     recent_activity,
	}
}

type activityTimelineItem struct {
	Date      string
	DayName   string
	Type      string
	Mood      *string
	Timestamp time.Time
}

func computeStreaks(successDates []time.Time, todayUTC time.Time) (current int, longest int) {
	if len(successDates) == 0 {
		return 0, 0
	}

	longest = 1
	currentRun := 1
	for i := 1; i < len(successDates); i++ {
		diff := int(successDates[i].Sub(successDates[i-1]).Hours() / 24)
		if diff == 1 {
			currentRun++
		} else {
			if currentRun > longest {
				longest = currentRun
			}
			currentRun = 1
		}
	}
	if currentRun > longest {
		longest = currentRun
	}

	latest := successDates[len(successDates)-1]
	if latest.Before(todayUTC.AddDate(0, 0, -1)) {
		return 0, longest
	}

	current = 1
	for i := len(successDates) - 1; i > 0; i-- {
		diff := int(successDates[i].Sub(successDates[i-1]).Hours() / 24)
		if diff != 1 {
			break
		}
		current++
	}
	return current, longest
}

func computeProgressPayload(checkIns []models.CheckIn, todayUTC time.Time, window_days int) ProgressPayload {
	currentEnd := todayUTC
	currentStart := currentEnd.AddDate(0, 0, -(window_days - 1))
	previousEnd := currentStart.AddDate(0, 0, -1)
	previousStart := previousEnd.AddDate(0, 0, -(window_days - 1))

	currentSuccess := countSuccessfulCheckInsInRange(checkIns, currentStart, currentEnd)
	previousSuccess := countSuccessfulCheckInsInRange(checkIns, previousStart, previousEnd)
	delta := currentSuccess - previousSuccess
	deltaRate := 0.0
	if previousSuccess > 0 {
		deltaRate = roundRatio(float64(delta) / float64(previousSuccess))
	}

	return ProgressPayload{
		WindowDays:                 window_days,
		CurrentSuccessfulCheckins:  currentSuccess,
		PreviousSuccessfulCheckins: previousSuccess,
		Delta:                      delta,
		DeltaRate:                  deltaRate,
	}
}

func zeroProgressPayload(window_days int) ProgressPayload {
	return ProgressPayload{
		WindowDays:                 window_days,
		CurrentSuccessfulCheckins:  0,
		PreviousSuccessfulCheckins: 0,
		Delta:                      0,
		DeltaRate:                  0,
	}
}

func countSuccessfulCheckInsInRange(checkIns []models.CheckIn, startDate time.Time, endDate time.Time) int {
	count := 0
	for _, row := range checkIns {
		if !row.IsSuccessful {
			continue
		}
		day := utcDayStart(row.CheckInDate)
		if day.Before(startDate) || day.After(endDate) {
			continue
		}
		count++
	}
	return count
}

func buildMoodTrendPayload(moodByDay map[string]map[string]int, successByDay map[string]int, totalByDay map[string]int) []MoodTrendPayload {
	keys := make([]string, 0, len(moodByDay))
	for day := range moodByDay {
		keys = append(keys, day)
	}
	sort.Strings(keys)

	result := make([]MoodTrendPayload, 0, len(keys))
	for _, day := range keys {
		moodMap := moodByDay[day]
		dominant := "unknown"
		maxCount := -1
		for mood, count := range moodMap {
			if count > maxCount {
				dominant = mood
				maxCount = count
			}
		}

		result = append(result, MoodTrendPayload{
			Date:            day,
			DayName:         dayNameID(parseDayOrZero(day)),
			DominantMood:    dominant,
			SuccessfulRatio: safeRatio(successByDay[day], totalByDay[day]),
		})
	}
	return result
}

func (s *Service) buildRelapseSolution(ctx context.Context, userID string, mood string, commitment *string, relapseTrigger []string) RelapseSolutionPayload {
	fallback := buildFallbackRelapseSolution(mood, relapseTrigger, s.now())
	if s.advisor == nil {
		return fallback
	}

	response, err := s.advisor.GenerateRelapseSolution(ctx, userID, aimodule.RelapseSolutionRequest{
		Mood:           mood,
		RelapseTrigger: relapseTrigger,
		Commitment:     commitment,
	})
	if err != nil {
		return fallback
	}

	return RelapseSolutionPayload{
		Title:       response.Title,
		Analysis:    response.Analysis,
		Summary:     response.Summary,
		GeneratedAt: s.now().UTC().Format(time.RFC3339),
	}
}

func buildFallbackRelapseSolution(mood string, relapseTrigger []string, nowUTC time.Time) RelapseSolutionPayload {
	triggerText := strings.TrimSpace(strings.Join(relapseTrigger, ", "))
	if triggerText == "" {
		triggerText = "pemicu belum dicatat"
	}
	return RelapseSolutionPayload{
		Title:       "Langkah Pemulihan Cepat",
		Analysis:    "Relapse terdeteksi dengan mood " + strings.TrimSpace(strings.ToLower(mood)) + ". Pola pemicu saat ini: " + triggerText + ".",
		Summary:     "Solusi terbaik saat ini: putus akses ke pemicu sekarang, stabilkan emosi singkat, lalu pindah ke aktivitas aman yang sudah disiapkan.",
		GeneratedAt: nowUTC.UTC().Format(time.RFC3339),
	}
}

func (s *Service) getCurrentAISummary(ctx context.Context, userID string) (string, error) {
	profile, err := s.repo.FindProfileByUserID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return defaultRoutineAISummaryMessage, nil
		}
		return "", errs.New(errs.CodeInternalError, "Gagal membaca ringkasan AI pengguna", nil, err)
	}

	if profile.AISummary == nil {
		return defaultRoutineAISummaryMessage, nil
	}

	summary := strings.TrimSpace(*profile.AISummary)
	if summary == "" {
		return defaultRoutineAISummaryMessage, nil
	}
	return summary, nil
}

func computeRelapseHourStats(relapses []models.Relapse) ([]RelapseHourStatPayload, []int, int) {
	if len(relapses) == 0 {
		return []RelapseHourStatPayload{}, []int{}, 0
	}

	countByHour := make([]int, 24)
	for _, row := range relapses {
		hour := row.CreatedAt.UTC().Hour()
		if hour < 0 || hour > 23 {
			continue
		}
		countByHour[hour]++
	}

	peakCount := 0
	for hour := 0; hour < 24; hour++ {
		if countByHour[hour] > peakCount {
			peakCount = countByHour[hour]
		}
	}

	hourlyDistribution := make([]RelapseHourStatPayload, 0, len(relapses))
	peakHoursUTC := make([]int, 0, 24)
	for hour := 0; hour < 24; hour++ {
		count := countByHour[hour]
		if count <= 0 {
			continue
		}
		hourlyDistribution = append(hourlyDistribution, RelapseHourStatPayload{
			HourUTC:      hour,
			RelapseCount: count,
		})
		if count == peakCount {
			peakHoursUTC = append(peakHoursUTC, hour)
		}
	}

	return hourlyDistribution, peakHoursUTC, peakCount
}

func computeRelapseTriggerStats(relapses []models.Relapse) []RelapseTriggerStatPayload {
	if len(relapses) == 0 {
		return []RelapseTriggerStatPayload{}
	}

	type triggerCounter struct {
		Key   string
		Label string
		Count int
	}

	countByTrigger := map[string]int{}
	labelByTrigger := map[string]string{}
	for _, row := range relapses {
		for _, trigger := range row.RelapseTrigger {
			label := strings.TrimSpace(trigger)
			if label == "" {
				continue
			}
			key := strings.ToLower(label)
			countByTrigger[key]++
			if _, exists := labelByTrigger[key]; !exists {
				labelByTrigger[key] = label
			}
		}
	}

	if len(countByTrigger) == 0 {
		return []RelapseTriggerStatPayload{}
	}

	counters := make([]triggerCounter, 0, len(countByTrigger))
	for key, count := range countByTrigger {
		counters = append(counters, triggerCounter{
			Key:   key,
			Label: labelByTrigger[key],
			Count: count,
		})
	}

	sort.Slice(counters, func(i, j int) bool {
		if counters[i].Count != counters[j].Count {
			return counters[i].Count > counters[j].Count
		}
		return counters[i].Key < counters[j].Key
	})

	result := make([]RelapseTriggerStatPayload, 0, len(counters))
	for _, item := range counters {
		result = append(result, RelapseTriggerStatPayload{
			RelapseTrigger:      item.Label,
			RelapseTriggerCount: item.Count,
		})
	}
	return result
}

func latestRelapseEvent(relapses []models.Relapse) models.Relapse {
	if len(relapses) == 0 {
		return models.Relapse{}
	}
	latest := relapses[0]
	for i := 1; i < len(relapses); i++ {
		candidate := relapses[i]
		if candidate.CreatedAt.UTC().After(latest.CreatedAt.UTC()) {
			latest = candidate
			continue
		}
		if candidate.CreatedAt.UTC().Equal(latest.CreatedAt.UTC()) && candidate.RelapseDate.UTC().After(latest.RelapseDate.UTC()) {
			latest = candidate
		}
	}
	return latest
}

func (s *Service) buildRelapseTimeSummary(ctx context.Context, userID string, relapses []models.Relapse, peakHoursUTC []int, peakCount int, latestRelapse models.Relapse) RelapseTimeSummaryPayload {
	if len(peakHoursUTC) == 0 {
		return emptyRelapseTimeSummary(s.now())
	}

	syntheticTrigger := buildPeakHourSyntheticTrigger(peakHoursUTC, peakCount, latestRelapse.RelapseTrigger)
	topTrigger, topTriggerCount := findMostFrequentRelapseTrigger(relapses)
	if topTrigger != "" {
		topTriggerLine := "trigger paling sering: " + topTrigger
		if topTriggerCount > 0 {
			topTriggerLine += " (" + formatInt(topTriggerCount) + " kejadian)"
		}
		syntheticTrigger = append(syntheticTrigger, topTriggerLine)
	}

	plan := s.buildRelapseSolution(ctx, userID, latestRelapse.Mood, latestRelapse.Commitment, syntheticTrigger)
	summary := strings.TrimSpace(plan.Summary)
	if topTrigger != "" {
		summary = "Trigger paling sering saat ini: " + topTrigger + ". " + summary
	}
	return RelapseTimeSummaryPayload{
		Title:       "Analisis Waktu Relapse",
		Analysis:    plan.Analysis,
		Summary:     summary,
		GeneratedAt: plan.GeneratedAt,
	}
}

func buildPeakHourSyntheticTrigger(peakHoursUTC []int, peakCount int, relapseTrigger []string) []string {
	result := make([]string, 0, len(relapseTrigger)+1)
	result = append(result, relapseTrigger...)

	formattedHours := make([]string, 0, len(peakHoursUTC))
	for _, hour := range peakHoursUTC {
		formattedHours = append(formattedHours, formatUTCHour(hour))
	}

	peakLine := "jam relapse paling sering (UTC): " + strings.Join(formattedHours, ", ")
	if peakCount > 0 {
		peakLine += " (" + formatInt(peakCount) + " kejadian)"
	}
	result = append(result, peakLine)
	return result
}

func emptyRelapseTimeSummary(nowUTC time.Time) RelapseTimeSummaryPayload {
	return RelapseTimeSummaryPayload{
		Title:       "Pola Relapse Belum Tersedia",
		Analysis:    "Belum ada data relapse untuk dianalisis. Gunakan UTC di backend, konversi jam di frontend sesuai zona user.",
		Summary:     "Solusi terbaik saat ini: mulai catat trigger konsisten agar analisis pola dan strategi pencegahan bisa dipersonalisasi.",
		GeneratedAt: nowUTC.UTC().Format(time.RFC3339),
	}
}

func findMostFrequentRelapseTrigger(relapses []models.Relapse) (string, int) {
	distribution := computeRelapseTriggerStats(relapses)
	if len(distribution) == 0 {
		return "", 0
	}
	return distribution[0].RelapseTrigger, distribution[0].RelapseTriggerCount
}

func formatUTCHour(hour int) string {
	if hour < 0 {
		hour = 0
	}
	if hour > 23 {
		hour = 23
	}
	if hour < 10 {
		return "0" + formatInt(hour) + ":00"
	}
	return formatInt(hour) + ":00"
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}

func buildWeekdaySummary(checkIns []models.CheckIn, relapses []models.Relapse) []WeekdaySummaryPayload {
	type dayCounter struct {
		success int
		relapse int
		total   int
	}

	counters := map[time.Weekday]dayCounter{}
	legacyRelapseDays := map[string]struct{}{}
	for _, row := range checkIns {
		day := utcDayStart(row.CheckInDate).Weekday()
		dayKey := utcDayStart(row.CheckInDate).Format("2006-01-02")
		counter := counters[day]
		counter.total++
		if row.IsSuccessful {
			counter.success++
		} else {
			counter.relapse++
			legacyRelapseDays[dayKey] = struct{}{}
		}
		counters[day] = counter
	}
	for _, relapse := range relapses {
		dayKey := utcDayStart(relapse.RelapseDate).Format("2006-01-02")
		if _, fromLegacy := legacyRelapseDays[dayKey]; fromLegacy {
			continue
		}
		day := utcDayStart(relapse.RelapseDate).Weekday()
		counter := counters[day]
		counter.total++
		counter.relapse++
		counters[day] = counter
	}

	orderedDays := []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
		time.Sunday,
	}
	result := make([]WeekdaySummaryPayload, 0, len(orderedDays))
	for _, day := range orderedDays {
		counter := counters[day]
		result = append(result, WeekdaySummaryPayload{
			DayName:            dayNameFromWeekday(day),
			SuccessfulCheckins: counter.success,
			RelapseCount:       counter.relapse,
			TotalCheckins:      counter.total,
			SuccessRate:        safeRatio(counter.success, counter.total),
		})
	}
	return result
}

func buildStreakGoalComparison(pornFreeGoal *int, currentStreak int, longestStreak int) StreakGoalComparisonPayload {
	payload := StreakGoalComparisonPayload{
		PornFreeGoal:  nil,
		CurrentStreak: currentStreak,
		LongestStreak: longestStreak,
		GoalReached:   false,
		RemainingDays: nil,
		ProgressRate:  0,
	}

	if pornFreeGoal == nil || *pornFreeGoal <= 0 {
		return payload
	}

	goal := *pornFreeGoal
	payload.PornFreeGoal = &goal
	payload.GoalReached = currentStreak >= goal
	if payload.GoalReached {
		remaining := 0
		payload.RemainingDays = &remaining
		payload.ProgressRate = 1
		return payload
	}

	remaining := goal - currentStreak
	if remaining < 0 {
		remaining = 0
	}
	payload.RemainingDays = &remaining
	payload.ProgressRate = roundRatio(float64(currentStreak) / float64(goal))
	return payload
}

func maxDayPtr(current *time.Time, next time.Time) *time.Time {
	day := utcDayStart(next)
	if current == nil || day.After(*current) {
		value := day
		return &value
	}
	return current
}

func mapLastDatePayload(date *time.Time) (*string, *string) {
	if date == nil {
		return nil, nil
	}
	formatted := utcDayStart(*date).Format("2006-01-02")
	dayName := dayNameID(*date)
	return &formatted, &dayName
}

func parseDayOrZero(raw string) time.Time {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func dayNameID(t time.Time) string {
	return dayNameFromWeekday(utcDayStart(t).Weekday())
}

func dayNameFromWeekday(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "Senin"
	case time.Tuesday:
		return "Selasa"
	case time.Wednesday:
		return "Rabu"
	case time.Thursday:
		return "Kamis"
	case time.Friday:
		return "Jumat"
	case time.Saturday:
		return "Sabtu"
	default:
		return "Minggu"
	}
}

func safeRatio(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return roundRatio(float64(numerator) / float64(denominator))
}

func roundRatio(value float64) float64 {
	return math.Round(value*100) / 100
}

func sameUTCDate(a time.Time, b time.Time) bool {
	ua := utcDayStart(a)
	ub := utcDayStart(b)
	return ua.Equal(ub)
}

func utcDayStart(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
