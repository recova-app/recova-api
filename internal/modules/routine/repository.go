package routine

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
)

const uniqueViolationCode = "23505"

// Repository provides persistence operations for routine module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs routine repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB exposes underlying GORM handle.
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// WithTx clones repository bound to transaction handle.
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// CloneTx clones repository bound to transaction handle as abstraction-friendly type.
func (r *Repository) CloneTx(tx *gorm.DB) routineRepository {
	return r.WithTx(tx)
}

// FindUserByID checks user existence by identifier.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindProfileByUserID reads profile row by user ID.
func (r *Repository) FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error) {
	var profile models.Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", strings.TrimSpace(userID)).First(&profile).Error; err != nil {
		return models.Profile{}, err
	}
	return profile, nil
}

// FindCheckInByUserAndDate loads check-in by user and UTC date.
func (r *Repository) FindCheckInByUserAndDate(ctx context.Context, userID string, check_in_date time.Time) (models.CheckIn, error) {
	var check_in models.CheckIn
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("check_in_date = ?", check_in_date.UTC()).
		First(&check_in).Error
	if err != nil {
		return models.CheckIn{}, err
	}
	return check_in, nil
}

// CreateCheckIn inserts a check-in row.
func (r *Repository) CreateCheckIn(ctx context.Context, check_in models.CheckIn) error {
	return r.db.WithContext(ctx).Create(&check_in).Error
}

// CreateJournal inserts journal row linked with user/check-in.
func (r *Repository) CreateJournal(ctx context.Context, journal models.Journal) error {
	return r.db.WithContext(ctx).Create(&journal).Error
}

// UpsertJournalByCheckInID creates or updates journal content linked to check-in.
func (r *Repository) UpsertJournalByCheckInID(ctx context.Context, userID string, checkInID string, content string) error {
	updates := r.db.WithContext(ctx).
		Model(&models.Journal{}).
		Where("check_in_id = ?", strings.TrimSpace(checkInID)).
		Update("content", strings.TrimSpace(content))
	if updates.Error != nil {
		return updates.Error
	}
	if updates.RowsAffected > 0 {
		return nil
	}

	checkInIDCopy := strings.TrimSpace(checkInID)
	return r.db.WithContext(ctx).Create(&models.Journal{
		UserID:    strings.TrimSpace(userID),
		CheckInID: &checkInIDCopy,
		Content:   strings.TrimSpace(content),
	}).Error
}

// FindActiveStreak returns active streak for user.
func (r *Repository) FindActiveStreak(ctx context.Context, userID string) (models.Streak, error) {
	var streak models.Streak
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("is_active = ?", true).
		Order("start_date desc").
		First(&streak).Error
	if err != nil {
		return models.Streak{}, err
	}
	return streak, nil
}

// CreateStreak inserts streak row.
func (r *Repository) CreateStreak(ctx context.Context, streak models.Streak) error {
	return r.db.WithContext(ctx).Create(&streak).Error
}

// CloseActiveStreak marks active streak as closed.
func (r *Repository) CloseActiveStreak(ctx context.Context, streakID string, endDate time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Streak{}).
		Where("id = ?", strings.TrimSpace(streakID)).
		Updates(map[string]any{
			"end_date":  endDate.UTC(),
			"is_active": false,
		}).Error
}

// LatestSuccessfulCheckInBeforeDate returns latest successful check-in date before target date.
func (r *Repository) LatestSuccessfulCheckInBeforeDate(ctx context.Context, userID string, targetDate time.Time) (*time.Time, error) {
	var check_in models.CheckIn
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("is_successful = ?", true).
		Where("check_in_date < ?", targetDate.UTC()).
		Order("check_in_date desc").
		First(&check_in).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	value := check_in.CheckInDate.UTC()
	return &value, nil
}

// ListSuccessfulCheckInsByUser returns successful check-ins sorted by day.
func (r *Repository) ListSuccessfulCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error) {
	var rows []models.CheckIn
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("is_successful = ?", true).
		Order("check_in_date asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListCheckInsByUser returns all check-ins sorted by day and creation time.
func (r *Repository) ListCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error) {
	var rows []models.CheckIn
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Order("check_in_date asc").
		Order("created_at asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListCheckInsByUserWithinDateRange returns check-ins within inclusive UTC date range.
func (r *Repository) ListCheckInsByUserWithinDateRange(ctx context.Context, userID string, startDate time.Time, endDate time.Time) ([]models.CheckIn, error) {
	var rows []models.CheckIn
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("check_in_date >= ?", startDate.UTC()).
		Where("check_in_date <= ?", endDate.UTC()).
		Order("check_in_date desc").
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListRelapseCheckInsByUser returns failed check-ins with optional journal linked by check-in.
func (r *Repository) ListRelapseCheckInsByUser(ctx context.Context, userID string) ([]models.CheckIn, error) {
	var rows []models.CheckIn
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("is_successful = ?", false).
		Order("check_in_date desc").
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// FindRelapseByUserAndDate loads relapse by user and UTC date.
func (r *Repository) FindRelapseByUserAndDate(ctx context.Context, userID string, relapseDate time.Time) (models.Relapse, error) {
	var relapse models.Relapse
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("relapse_date = ?", relapseDate.UTC()).
		First(&relapse).Error
	if err != nil {
		return models.Relapse{}, err
	}
	return relapse, nil
}

// CreateRelapse inserts relapse row.
func (r *Repository) CreateRelapse(ctx context.Context, relapse models.Relapse) error {
	return r.db.WithContext(ctx).Create(&relapse).Error
}

// ListRelapsesByUser returns all relapse rows sorted by day and creation time.
func (r *Repository) ListRelapsesByUser(ctx context.Context, userID string) ([]models.Relapse, error) {
	var rows []models.Relapse
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Order("relapse_date desc").
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListRelapsesByUserWithinDateRange returns relapses within inclusive UTC date range.
func (r *Repository) ListRelapsesByUserWithinDateRange(ctx context.Context, userID string, startDate time.Time, endDate time.Time) ([]models.Relapse, error) {
	var rows []models.Relapse
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("relapse_date >= ?", startDate.UTC()).
		Where("relapse_date <= ?", endDate.UTC()).
		Order("relapse_date desc").
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// FindJournalsByCheckInIDs maps journal rows using linked check-in IDs.
func (r *Repository) FindJournalsByCheckInIDs(ctx context.Context, checkInIDs []string) ([]models.Journal, error) {
	if len(checkInIDs) == 0 {
		return []models.Journal{}, nil
	}

	var rows []models.Journal
	if err := r.db.WithContext(ctx).
		Where("check_in_id IN ?", checkInIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListJournalsByUserWithinTimeRange returns journals within inclusive start and exclusive end time.
func (r *Repository) ListJournalsByUserWithinTimeRange(ctx context.Context, userID string, startAt time.Time, endAt time.Time) ([]models.Journal, error) {
	var rows []models.Journal
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("created_at >= ?", startAt.UTC()).
		Where("created_at < ?", endAt.UTC()).
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueViolation reports postgres unique violation errors.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolationCode
}
