package routine

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
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
	FindCheckInByUserAndDate(ctx context.Context, userID string, check_in_date time.Time) (models.CheckIn, error)
	CreateCheckIn(ctx context.Context, check_in models.CheckIn) error
	CreateJournal(ctx context.Context, journal models.Journal) error
	FindActiveStreak(ctx context.Context, userID string) (models.Streak, error)
	CreateStreak(ctx context.Context, streak models.Streak) error
	CloseActiveStreak(ctx context.Context, streakID string, endDate time.Time) error
	LatestSuccessfulCheckInBeforeDate(ctx context.Context, userID string, targetDate time.Time) (*time.Time, error)
	ListSuccessfulCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
	ListCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
	ListCheckInsByUserWithinDateRange(ctx context.Context, userID string, startDate time.Time, endDate time.Time) ([]models.CheckIn, error)
	ListRelapseCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
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
			RelapseTrigger: pq.StringArray(append([]string{}, input.RelapseTrigger...)),
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

	var relapseSolution *RelapseSolutionPayload
	if !input.IsSuccessful {
		solution := s.buildRelapseSolution(ctx, userID, input)
		relapseSolution = &solution
	}

	return CheckInResponseData{
		CheckIn:         mapCheckInPayload(stored, input.JournalText),
		Statistics:      stats,
		RelapseSolution: relapseSolution,
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

	return computeStatistics(rows, utcDayStart(s.now()), user.PornFreeGoal), nil
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

	endExclusive := endDate.AddDate(0, 0, 1)
	journals, err := s.repo.ListJournalsByUserWithinTimeRange(ctx, userID, startDate, endExclusive)
	if err != nil {
		return ActivitySummaryPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca ringkasan aktivitas", nil, err)
	}

	return computeActivitySummary(window_days, checkIns, journals), nil
}

// GetRelapses returns user relapse history (failed check-ins).
func (s *Service) GetRelapses(ctx context.Context, userID string) ([]RelapsePayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return nil, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListRelapseCheckInsByUser(ctx, userID)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca riwayat relapse", nil, err)
	}

	checkInIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		checkInIDs = append(checkInIDs, row.ID)
	}

	journals, err := s.repo.FindJournalsByCheckInIDs(ctx, checkInIDs)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca catatan relapse", nil, err)
	}

	journalByCheckIn := map[string]string{}
	for _, journal := range journals {
		if journal.CheckInID == nil {
			continue
		}
		journalByCheckIn[strings.TrimSpace(*journal.CheckInID)] = journal.Content
	}

	result := make([]RelapsePayload, 0, len(rows))
	for _, row := range rows {
		commitment := journalByCheckIn[strings.TrimSpace(row.ID)]
		var commitmentPtr *string
		if strings.TrimSpace(commitment) != "" {
			value := commitment
			commitmentPtr = &value
		}

		result = append(result, RelapsePayload{
			CheckInID:      row.ID,
			CheckInDate:    row.CheckInDate.UTC().Format("2006-01-02"),
			CheckInDayName: dayNameID(row.CheckInDate),
			Mood:           row.Mood,
			Commitment:     commitmentPtr,
			RelapseTrigger: append([]string{}, row.RelapseTrigger...),
			CreatedAt:      row.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return result, nil
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

	previousSuccessfulDate, err := repo.LatestSuccessfulCheckInBeforeDate(ctx, userID, check_in_date)
	if err != nil {
		return errs.New(errs.CodeInternalError, "Gagal membaca histori check-in", nil, err)
	}

	if err != nil || IsRecordNotFound(err) {
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

func computeStatistics(checkIns []models.CheckIn, todayUTC time.Time, pornFreeGoal *int) StatisticsPayload {
	if len(checkIns) == 0 {
		return StatisticsPayload{
			CurrentStreak:           0,
			LongestStreak:           0,
			TotalCheckins:           0,
			TotalAttempts:           0,
			SuccessRate:             0,
			StreakCalendar:          []string{},
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
			WeekdaySummary:          buildWeekdaySummary(nil),
			StreakGoalComparison:    buildStreakGoalComparison(pornFreeGoal, 0, 0),
		}
	}

	successDates := make([]time.Time, 0, len(checkIns))
	calendar := make([]string, 0, len(checkIns))
	active_days := map[string]struct{}{}
	successCount := 0
	relapse_count := 0
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
		lastCheckInDate = &day

		if _, exists := moodByDay[dayKey]; !exists {
			moodByDay[dayKey] = map[string]int{}
		}
		moodKey := strings.TrimSpace(strings.ToLower(row.Mood))
		if moodKey == "" {
			moodKey = "unknown"
		}
		moodByDay[dayKey][moodKey]++

		if row.IsSuccessful {
			successCount++
			successCountByDay[dayKey]++
			successDates = append(successDates, day)
			calendar = append(calendar, dayKey)
			continue
		}
		relapse_count++
		lastRelapseDate = &day
	}

	totalAttempts := successCount + relapse_count
	recovery_success_rate := safeRatio(successCount, totalAttempts)
	relapse_rate := safeRatio(relapse_count, totalAttempts)
	success_rate := safeRatio(successCount, totalAttempts)

	current_streak, longest_streak := computeStreaks(successDates, todayUTC)
	checkin_consistency_score := safeRatio(len(active_days), 30)
	weekly_progress := computeProgressPayload(checkIns, todayUTC, 7)
	monthly_progress := computeProgressPayload(checkIns, todayUTC, 30)
	mood_trend := buildMoodTrendPayload(moodByDay, successCountByDay, totalCountByDay)
	weekdaySummary := buildWeekdaySummary(checkIns)
	lastCheckInDateRaw, lastCheckInDayName := mapLastDatePayload(lastCheckInDate)
	lastRelapseDateRaw, lastRelapseDayName := mapLastDatePayload(lastRelapseDate)

	return StatisticsPayload{
		CurrentStreak:           current_streak,
		LongestStreak:           longest_streak,
		TotalCheckins:           successCount,
		TotalAttempts:           totalAttempts,
		SuccessRate:             success_rate,
		StreakCalendar:          calendar,
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

func computeActivitySummary(window_days int, checkIns []models.CheckIn, journals []models.Journal) ActivitySummaryPayload {
	successful_checkins := 0
	relapses := 0
	active_days := map[string]struct{}{}
	checkinByID := make(map[string]models.CheckIn, len(checkIns))
	activities := make([]activityTimelineItem, 0, len(checkIns)+len(journals))

	for _, row := range checkIns {
		day := utcDayStart(row.CheckInDate).Format("2006-01-02")
		active_days[day] = struct{}{}
		checkinByID[strings.TrimSpace(row.ID)] = row

		if row.IsSuccessful {
			successful_checkins++
		} else {
			relapses++
		}

		eventType := "checkin_relapse"
		if row.IsSuccessful {
			eventType = "checkin_success"
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
		Relapses:           relapses,
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

func (s *Service) buildRelapseSolution(ctx context.Context, userID string, input DailyCheckInInput) RelapseSolutionPayload {
	fallback := buildFallbackRelapseSolution(input.Mood, input.RelapseTrigger, s.now())
	if s.advisor == nil {
		return fallback
	}

	response, err := s.advisor.GenerateRelapseSolution(ctx, userID, aimodule.RelapseSolutionRequest{
		Mood:           input.Mood,
		RelapseTrigger: input.RelapseTrigger,
		Commitment:     input.JournalText,
	})
	if err != nil {
		return fallback
	}

	return RelapseSolutionPayload{
		Title:       response.Title,
		Analysis:    response.Analysis,
		ActionSteps: append([]string{}, response.ActionSteps...),
		GeneratedAt: s.now().UTC().Format(time.RFC3339),
	}
}

func buildFallbackRelapseSolution(mood string, relapseTrigger []string, nowUTC time.Time) RelapseSolutionPayload {
	triggerText := strings.TrimSpace(strings.Join(relapseTrigger, ", "))
	if triggerText == "" {
		triggerText = "pemicu belum dicatat"
	}
	return RelapseSolutionPayload{
		Title:    "Langkah Pemulihan Cepat",
		Analysis: "Relapse terdeteksi dengan mood " + strings.TrimSpace(strings.ToLower(mood)) + ". Fokus dulu stabilkan emosi dan putus rantai pemicu.",
		ActionSteps: []string{
			"Tarik napas dalam 4 kali, lalu jauhkan diri dari pemicu selama 10 menit.",
			"Tulis 1 kalimat: pemicu utama hari ini = " + triggerText + ".",
			"Hubungi support system atau buka konten pemulihan sebelum kembali ke aktivitas.",
		},
		GeneratedAt: nowUTC.UTC().Format(time.RFC3339),
	}
}

func buildWeekdaySummary(checkIns []models.CheckIn) []WeekdaySummaryPayload {
	type dayCounter struct {
		success int
		relapse int
		total   int
	}

	counters := map[time.Weekday]dayCounter{}
	for _, row := range checkIns {
		day := utcDayStart(row.CheckInDate).Weekday()
		counter := counters[day]
		counter.total++
		if row.IsSuccessful {
			counter.success++
		} else {
			counter.relapse++
		}
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
