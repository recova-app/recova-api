package achievements

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

type achievementsRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	ListActiveAchievements(ctx context.Context, category *string) ([]models.Achievement, error)
	ListProgressByUser(ctx context.Context, userID string, category *string) ([]progressListRow, error)
	ListUnlockedByUser(ctx context.Context, userID string, category *string) ([]unlockedListRow, error)
	ComputeEvaluationMetrics(ctx context.Context, userID string, nowUTC time.Time) (evaluationMetrics, error)
	UpsertProgress(ctx context.Context, userID string, rows []progressUpsert, evaluatedAt time.Time) error
}

// Service owns achievements business rules.
type Service struct {
	repo achievementsRepository
	now  func() time.Time
}

// NewService constructs achievements service.
func NewService(repo achievementsRepository) *Service {
	return &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// GetCatalog returns active achievement catalog payload.
func (s *Service) GetCatalog(ctx context.Context, userID string, query CategoryQuery) (CatalogResponse, error) {
	category, err := NormalizeCategoryQuery(query.Category)
	if err != nil {
		return CatalogResponse{}, err
	}

	if err := s.ensureUserExists(ctx, userID); err != nil {
		return CatalogResponse{}, err
	}

	rows, err := s.repo.ListActiveAchievements(ctx, category)
	if err != nil {
		return CatalogResponse{}, errs.New(errs.CodeInternalError, "Gagal membaca katalog achievement", nil, err)
	}

	items := make([]CatalogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, CatalogItem{
			ID:          row.ID,
			Code:        row.Code,
			Title:       row.Title,
			Description: row.Description,
			Category:    row.Category,
			Threshold:   row.Threshold,
		})
	}

	return CatalogResponse{Items: items}, nil
}

// GetProgress evaluates and returns user achievement progress payload.
func (s *Service) GetProgress(ctx context.Context, userID string, query CategoryQuery) (ProgressResponse, error) {
	category, err := NormalizeCategoryQuery(query.Category)
	if err != nil {
		return ProgressResponse{}, err
	}

	if err := s.ensureUserExists(ctx, userID); err != nil {
		return ProgressResponse{}, err
	}

	if err := s.evaluateAndPersist(ctx, userID); err != nil {
		return ProgressResponse{}, err
	}

	rows, err := s.repo.ListProgressByUser(ctx, userID, category)
	if err != nil {
		return ProgressResponse{}, errs.New(errs.CodeInternalError, "Gagal membaca progres achievement", nil, err)
	}

	items := make([]ProgressItem, 0, len(rows))
	for _, row := range rows {
		var unlockedAt *string
		if row.UnlockedAt != nil {
			value := row.UnlockedAt.UTC().Format(time.RFC3339)
			unlockedAt = &value
		}
		items = append(items, ProgressItem{
			AchievementCode: row.Code,
			Category:        row.Category,
			Threshold:       row.Threshold,
			ProgressValue:   row.ProgressValue,
			Unlocked:        row.UnlockedAt != nil,
			UnlockedAt:      unlockedAt,
		})
	}

	return ProgressResponse{Items: items}, nil
}

// GetUnlocked evaluates and returns unlocked achievements payload.
func (s *Service) GetUnlocked(ctx context.Context, userID string, query CategoryQuery) (UnlockedResponse, error) {
	category, err := NormalizeCategoryQuery(query.Category)
	if err != nil {
		return UnlockedResponse{}, err
	}

	if err := s.ensureUserExists(ctx, userID); err != nil {
		return UnlockedResponse{}, err
	}

	if err := s.evaluateAndPersist(ctx, userID); err != nil {
		return UnlockedResponse{}, err
	}

	rows, err := s.repo.ListUnlockedByUser(ctx, userID, category)
	if err != nil {
		return UnlockedResponse{}, errs.New(errs.CodeInternalError, "Gagal membaca achievement yang telah terbuka", nil, err)
	}

	items := make([]UnlockedItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, UnlockedItem{
			AchievementCode: row.Code,
			Title:           row.Title,
			Description:     row.Description,
			Category:        row.Category,
			Threshold:       row.Threshold,
			ProgressValue:   row.ProgressValue,
			UnlockedAt:      row.UnlockedAt.UTC().Format(time.RFC3339),
		})
	}

	return UnlockedResponse{Items: items}, nil
}

func (s *Service) ensureUserExists(ctx context.Context, userID string) error {
	if _, err := s.repo.FindUserByID(ctx, strings.TrimSpace(userID)); err != nil {
		if IsRecordNotFound(err) {
			return errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}
	return nil
}

func (s *Service) evaluateAndPersist(ctx context.Context, userID string) error {
	achievements, err := s.repo.ListActiveAchievements(ctx, nil)
	if err != nil {
		return errs.New(errs.CodeInternalError, "Gagal membaca katalog achievement", nil, err)
	}
	if len(achievements) == 0 {
		return nil
	}

	nowUTC := s.now().UTC()
	metrics, err := s.repo.ComputeEvaluationMetrics(ctx, strings.TrimSpace(userID), nowUTC)
	if err != nil {
		return errs.New(errs.CodeInternalError, "Gagal mengevaluasi progres achievement", nil, err)
	}

	rows := make([]progressUpsert, 0, len(achievements))
	for _, achievement := range achievements {
		progressValue := progressForCategory(strings.TrimSpace(achievement.Category), metrics)
		var unlockedAt *time.Time
		if progressValue >= achievement.Threshold {
			t := nowUTC
			unlockedAt = &t
		}
		rows = append(rows, progressUpsert{
			AchievementID: achievement.ID,
			ProgressValue: progressValue,
			UnlockedAt:    unlockedAt,
		})
	}

	if err := s.repo.UpsertProgress(ctx, strings.TrimSpace(userID), rows, nowUTC); err != nil {
		return errs.New(errs.CodeInternalError, "Gagal menyimpan progres achievement", nil, err)
	}

	return nil
}

func progressForCategory(category string, metrics evaluationMetrics) float64 {
	switch strings.TrimSpace(category) {
	case categoryStreakMilestone:
		return metrics.StreakMilestone
	case categoryCheckinConsistency:
		return metrics.CheckinConsistency
	case categoryJournalConsistency:
		return metrics.JournalConsistency
	case categoryRelapseRecovery:
		return metrics.RelapseRecovery
	case categoryCommunityParticipation:
		return metrics.CommunityParticipation
	case categoryOnboardingCompletion:
		return metrics.OnboardingCompletion
	default:
		return 0
	}
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
