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

// FindCheckInByUserAndDate loads check-in by user and UTC date.
func (r *Repository) FindCheckInByUserAndDate(ctx context.Context, userID string, checkInDate time.Time) (models.CheckIn, error) {
	var checkIn models.CheckIn
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("check_in_date = ?", checkInDate.UTC()).
		First(&checkIn).Error
	if err != nil {
		return models.CheckIn{}, err
	}
	return checkIn, nil
}

// CreateCheckIn inserts a check-in row.
func (r *Repository) CreateCheckIn(ctx context.Context, checkIn models.CheckIn) error {
	return r.db.WithContext(ctx).Create(&checkIn).Error
}

// CreateJournal inserts journal row linked with user/check-in.
func (r *Repository) CreateJournal(ctx context.Context, journal models.Journal) error {
	return r.db.WithContext(ctx).Create(&journal).Error
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
	var checkIn models.CheckIn
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("is_successful = ?", true).
		Where("check_in_date < ?", targetDate.UTC()).
		Order("check_in_date desc").
		First(&checkIn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	value := checkIn.CheckInDate.UTC()
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
