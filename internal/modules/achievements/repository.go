package achievements

import (
	"context"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const rollingWindowDays = 30

// Repository provides persistence operations for achievements module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs achievements repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByID checks user existence by identifier.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// ListActiveAchievements loads active achievement catalog, optionally filtered by category.
func (r *Repository) ListActiveAchievements(ctx context.Context, category *string) ([]models.Achievement, error) {
	query := r.db.WithContext(ctx).
		Model(&models.Achievement{}).
		Where("is_active = ?", true).
		Order("category asc").
		Order("threshold asc").
		Order("code asc")
	if category != nil {
		query = query.Where("category = ?", strings.TrimSpace(*category))
	}

	var rows []models.Achievement
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type progressListRow struct {
	ID            string
	Code          string
	Title         string
	Description   string
	Category      string
	Threshold     float64
	ProgressValue float64
	UnlockedAt    *time.Time
}

// ListProgressByUser loads achievement progress rows for one user.
func (r *Repository) ListProgressByUser(ctx context.Context, userID string, category *string) ([]progressListRow, error) {
	query := r.db.WithContext(ctx).
		Table("achievements AS a").
		Select(`
			a.id,
			a.code,
			a.title,
			a.description,
			a.category,
			a.threshold,
			COALESCE(up.progress_value, 0) AS progress_value,
			up.unlocked_at
		`).
		Joins(`
			LEFT JOIN user_achievement_progress AS up
				ON up.achievement_id = a.id
				AND up.user_id = ?
		`, strings.TrimSpace(userID)).
		Where("a.is_active = ?", true).
		Order("a.category asc").
		Order("a.threshold asc").
		Order("a.code asc")
	if category != nil {
		query = query.Where("a.category = ?", strings.TrimSpace(*category))
	}

	var rows []progressListRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type unlockedListRow struct {
	Code          string
	Title         string
	Description   string
	Category      string
	Threshold     float64
	ProgressValue float64
	UnlockedAt    time.Time
}

// ListUnlockedByUser loads unlocked achievements for one user.
func (r *Repository) ListUnlockedByUser(ctx context.Context, userID string, category *string) ([]unlockedListRow, error) {
	query := r.db.WithContext(ctx).
		Table("user_achievement_progress AS up").
		Select(`
			a.code,
			a.title,
			a.description,
			a.category,
			a.threshold,
			up.progress_value,
			up.unlocked_at
		`).
		Joins("JOIN achievements AS a ON a.id = up.achievement_id").
		Where("up.user_id = ?", strings.TrimSpace(userID)).
		Where("up.unlocked_at IS NOT NULL").
		Where("a.is_active = ?", true).
		Order("up.unlocked_at desc").
		Order("a.code asc")
	if category != nil {
		query = query.Where("a.category = ?", strings.TrimSpace(*category))
	}

	var rows []unlockedListRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type evaluationMetrics struct {
	StreakMilestone        float64
	CheckinConsistency     float64
	JournalConsistency     float64
	RelapseRecovery        float64
	CommunityParticipation float64
	OnboardingCompletion   float64
}

// ComputeEvaluationMetrics calculates progress metrics for one user using batched queries.
func (r *Repository) ComputeEvaluationMetrics(ctx context.Context, userID string, nowUTC time.Time) (evaluationMetrics, error) {
	userID = strings.TrimSpace(userID)
	dayStart := utcDayStart(nowUTC)
	windowStart := dayStart.AddDate(0, 0, -(rollingWindowDays - 1))
	windowEndExclusive := dayStart.AddDate(0, 0, 1)

	streakProgress, err := r.currentStreakDays(ctx, userID, dayStart)
	if err != nil {
		return evaluationMetrics{}, err
	}
	relapseRecovery, err := r.relapseRecoveryScore(ctx, userID)
	if err != nil {
		return evaluationMetrics{}, err
	}
	activity, err := r.batchActivityMetrics(ctx, userID, windowStart, dayStart, windowEndExclusive)
	if err != nil {
		return evaluationMetrics{}, err
	}

	onboardingScore := 0.0
	if activity.OnboardingDone > 0 {
		onboardingScore = 1.0
	}

	return evaluationMetrics{
		StreakMilestone:        streakProgress,
		CheckinConsistency:     float64(activity.CheckinCount),
		JournalConsistency:     float64(activity.JournalCount),
		RelapseRecovery:        relapseRecovery,
		CommunityParticipation: float64(activity.CommunityCount),
		OnboardingCompletion:   onboardingScore,
	}, nil
}

type progressUpsert struct {
	AchievementID string
	ProgressValue float64
	UnlockedAt    *time.Time
}

// UpsertProgress stores evaluation result idempotently and prevents unlock regression.
func (r *Repository) UpsertProgress(ctx context.Context, userID string, rows []progressUpsert, evaluatedAt time.Time) error {
	if len(rows) == 0 {
		return nil
	}

	progressRows := make([]models.UserAchievementProgress, 0, len(rows))
	trimmedUserID := strings.TrimSpace(userID)
	for _, row := range rows {
		progressRows = append(progressRows, models.UserAchievementProgress{
			UserID:          trimmedUserID,
			AchievementID:   strings.TrimSpace(row.AchievementID),
			ProgressValue:   row.ProgressValue,
			UnlockedAt:      row.UnlockedAt,
			LastEvaluatedAt: evaluatedAt.UTC(),
			UpdatedAt:       evaluatedAt.UTC(),
		})
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "achievement_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"progress_value":    gorm.Expr("GREATEST(user_achievement_progress.progress_value, EXCLUDED.progress_value)"),
				"unlocked_at":       gorm.Expr("COALESCE(user_achievement_progress.unlocked_at, EXCLUDED.unlocked_at)"),
				"last_evaluated_at": gorm.Expr("EXCLUDED.last_evaluated_at"),
				"updated_at":        gorm.Expr("EXCLUDED.updated_at"),
			}),
		}).
		Create(&progressRows).Error
}

func (r *Repository) currentStreakDays(ctx context.Context, userID string, dayStart time.Time) (float64, error) {
	type streakRow struct {
		StartDate time.Time
	}

	var row streakRow
	err := r.db.WithContext(ctx).
		Table("streaks").
		Select("start_date").
		Where("user_id = ?", userID).
		Where("is_active = ?", true).
		Order("start_date desc").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	if row.StartDate.IsZero() {
		return 0, nil
	}

	start := utcDayStart(row.StartDate)
	if start.After(dayStart) {
		return 0, nil
	}

	days := int(dayStart.Sub(start).Hours()/24) + 1
	if days < 0 {
		days = 0
	}
	return float64(days), nil
}

type activityMetricsRow struct {
	CheckinCount   int64 `gorm:"column:checkin_count"`
	JournalCount   int64 `gorm:"column:journal_count"`
	OnboardingDone int64 `gorm:"column:onboarding_done"`
	CommunityCount int64 `gorm:"column:community_count"`
}

func (r *Repository) batchActivityMetrics(
	ctx context.Context,
	userID string,
	windowStart time.Time,
	dayStart time.Time,
	windowEndExclusive time.Time,
) (activityMetricsRow, error) {
	var row activityMetricsRow
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				(SELECT COUNT(*) FROM check_ins WHERE user_id = ? AND is_successful = true AND check_in_date >= ? AND check_in_date <= ?) AS checkin_count,
				(SELECT COUNT(*) FROM journals WHERE user_id = ? AND created_at >= ? AND created_at < ?) AS journal_count,
				(SELECT COUNT(*) FROM profiles WHERE user_id = ?) AS onboarding_done,
				(
					(SELECT COUNT(*) FROM community_posts WHERE user_id = ?) +
					(SELECT COUNT(*) FROM community_comments WHERE user_id = ?) +
					(SELECT COUNT(*) FROM community_post_likes WHERE user_id = ?)
				) AS community_count
		`, userID, windowStart, dayStart, userID, windowStart, windowEndExclusive, userID, userID, userID, userID).
		Scan(&row).Error
	if err != nil {
		return activityMetricsRow{}, err
	}
	return row, nil
}

func (r *Repository) relapseRecoveryScore(ctx context.Context, userID string) (float64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("check_ins").
		Where("user_id = ?", userID).
		Where("is_successful = ?", true).
		Where(`
			check_in_date > COALESCE(
				(SELECT check_in_date FROM check_ins WHERE user_id = ? AND is_successful = false ORDER BY check_in_date DESC LIMIT 1),
				'0001-01-01'::date
			)
		`, userID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return float64(count), nil
}

func utcDayStart(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
