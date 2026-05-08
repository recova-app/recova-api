package routine

import (
	"context"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

type routineRepository interface {
	DB() *gorm.DB
	CloneTx(tx *gorm.DB) routineRepository
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	FindCheckInByUserAndDate(ctx context.Context, userID string, checkInDate time.Time) (models.CheckIn, error)
	CreateCheckIn(ctx context.Context, checkIn models.CheckIn) error
	CreateJournal(ctx context.Context, journal models.Journal) error
	FindActiveStreak(ctx context.Context, userID string) (models.Streak, error)
	CreateStreak(ctx context.Context, streak models.Streak) error
	CloseActiveStreak(ctx context.Context, streakID string, endDate time.Time) error
	LatestSuccessfulCheckInBeforeDate(ctx context.Context, userID string, targetDate time.Time) (*time.Time, error)
	ListSuccessfulCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
	ListRelapseCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error)
	FindJournalsByCheckInIDs(ctx context.Context, checkInIDs []string) ([]models.Journal, error)
}

// Service owns routine business rules.
type Service struct {
	repo routineRepository
	now  func() time.Time
}

// NewService constructs routine service with repository dependency.
func NewService(repo routineRepository) *Service {
	return &Service{
		repo: repo,
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

	checkInDate := utcDayStart(s.now())

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
			UserID:       strings.TrimSpace(userID),
			CheckInDate:  checkInDate,
			Mood:         input.Mood,
			IsSuccessful: input.IsSuccessful,
		}
		if err := txRepo.CreateCheckIn(ctx, row); err != nil {
			if IsUniqueViolation(err) {
				return errs.New(errs.CodeConflict, "Pengguna sudah melakukan check-in hari ini", nil, err)
			}
			return errs.New(errs.CodeInternalError, "Gagal menyimpan check-in harian", nil, err)
		}

		created, err := txRepo.FindCheckInByUserAndDate(ctx, userID, checkInDate)
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

		if err := s.syncStreak(ctx, txRepo, userID, checkInDate, input.IsSuccessful); err != nil {
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
		CheckIn:    mapCheckInPayload(stored, input.JournalText),
		Statistics: stats,
	}, nil
}

// GetStatistics returns routine statistics for authenticated user.
func (s *Service) GetStatistics(ctx context.Context, userID string) (StatisticsPayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return StatisticsPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return StatisticsPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListSuccessfulCheckInsByUser(ctx, userID)
	if err != nil {
		return StatisticsPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca statistik check-in", nil, err)
	}

	return computeStatistics(rows, utcDayStart(s.now())), nil
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
			CheckInID:   row.ID,
			CheckInDate: row.CheckInDate.UTC().Format("2006-01-02"),
			Mood:        row.Mood,
			Commitment:  commitmentPtr,
			CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *Service) syncStreak(ctx context.Context, repo routineRepository, userID string, checkInDate time.Time, isSuccessful bool) error {
	active, err := repo.FindActiveStreak(ctx, userID)
	if err != nil && !IsRecordNotFound(err) {
		return errs.New(errs.CodeInternalError, "Gagal membaca status streak", nil, err)
	}

	if !isSuccessful {
		if err == nil {
			if closeErr := repo.CloseActiveStreak(ctx, active.ID, checkInDate); closeErr != nil {
				return errs.New(errs.CodeInternalError, "Gagal menutup streak aktif", nil, closeErr)
			}
		}
		return nil
	}

	previousSuccessfulDate, err := repo.LatestSuccessfulCheckInBeforeDate(ctx, userID, checkInDate)
	if err != nil {
		return errs.New(errs.CodeInternalError, "Gagal membaca histori check-in", nil, err)
	}

	if err != nil || IsRecordNotFound(err) {
		if createErr := repo.CreateStreak(ctx, models.Streak{
			UserID:    strings.TrimSpace(userID),
			StartDate: checkInDate,
			IsActive:  true,
		}); createErr != nil {
			return errs.New(errs.CodeInternalError, "Gagal membuat streak baru", nil, createErr)
		}
		return nil
	}

	if active.ID == "" {
		if createErr := repo.CreateStreak(ctx, models.Streak{
			UserID:    strings.TrimSpace(userID),
			StartDate: checkInDate,
			IsActive:  true,
		}); createErr != nil {
			return errs.New(errs.CodeInternalError, "Gagal membuat streak baru", nil, createErr)
		}
		return nil
	}

	yesterday := checkInDate.AddDate(0, 0, -1)
	if previousSuccessfulDate != nil && sameUTCDate(*previousSuccessfulDate, yesterday) {
		return nil
	}

	if err := repo.CloseActiveStreak(ctx, active.ID, checkInDate); err != nil {
		return errs.New(errs.CodeInternalError, "Gagal menutup streak aktif", nil, err)
	}
	if err := repo.CreateStreak(ctx, models.Streak{
		UserID:    strings.TrimSpace(userID),
		StartDate: checkInDate,
		IsActive:  true,
	}); err != nil {
		return errs.New(errs.CodeInternalError, "Gagal membuat streak baru", nil, err)
	}

	return nil
}

func mapCheckInPayload(row models.CheckIn, commitment *string) CheckInPayload {
	return CheckInPayload{
		ID:           row.ID,
		UserID:       row.UserID,
		CheckInDate:  row.CheckInDate.UTC().Format("2006-01-02"),
		Mood:         row.Mood,
		IsSuccessful: row.IsSuccessful,
		Commitment:   commitment,
		CreatedAt:    row.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func computeStatistics(checkIns []models.CheckIn, todayUTC time.Time) StatisticsPayload {
	total := len(checkIns)
	if total == 0 {
		return StatisticsPayload{
			CurrentStreak:  0,
			LongestStreak:  0,
			TotalCheckins:  0,
			StreakCalendar: []string{},
		}
	}

	dates := make([]time.Time, 0, len(checkIns))
	calendar := make([]string, 0, len(checkIns))
	for _, row := range checkIns {
		date := utcDayStart(row.CheckInDate)
		dates = append(dates, date)
		calendar = append(calendar, date.Format("2006-01-02"))
	}

	longest := 1
	currentRun := 1
	for i := 1; i < len(dates); i++ {
		diff := int(dates[i].Sub(dates[i-1]).Hours() / 24)
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

	current := 0
	latest := dates[len(dates)-1]
	if !latest.Before(todayUTC.AddDate(0, 0, -1)) {
		current = 1
		for i := len(dates) - 1; i > 0; i-- {
			diff := int(dates[i].Sub(dates[i-1]).Hours() / 24)
			if diff != 1 {
				break
			}
			current++
		}
	}

	return StatisticsPayload{
		CurrentStreak:  current,
		LongestStreak:  longest,
		TotalCheckins:  total,
		StreakCalendar: calendar,
	}
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
