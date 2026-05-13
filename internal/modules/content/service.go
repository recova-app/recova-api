package content

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

const (
	fallbackMotivation = "Teruslah maju, sekecil apapun langkahmu."
	fallbackChallenge  = "Tuliskan satu hal yang kamu syukuri hari ini."
)

type contentRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	ListActiveMotivations(ctx context.Context) ([]models.DailyMotivation, error)
	ListActiveChallenges(ctx context.Context) ([]models.DailyChallenge, error)
}

// Service owns daily content selection rules.
type Service struct {
	repo contentRepository
	now  func() time.Time
}

// NewService constructs content service.
func NewService(repo contentRepository) *Service {
	return &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// GetDailyContent returns deterministic daily motivation/challenge for authenticated user.
func (s *Service) GetDailyContent(ctx context.Context, userID string) (DailyContentPayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DailyContentPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return DailyContentPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	motivations, err := s.repo.ListActiveMotivations(ctx)
	if err != nil {
		return DailyContentPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca motivasi harian", nil, err)
	}
	challenges, err := s.repo.ListActiveChallenges(ctx)
	if err != nil {
		return DailyContentPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca tantangan harian", nil, err)
	}

	today := dayStartUTC(s.now())
	dateKey := today.Format("2006-01-02")

	motivation := fallbackMotivation
	if len(motivations) > 0 {
		idx := stableIndexForDate(today, len(motivations))
		candidate := strings.TrimSpace(motivations[idx].Content)
		if candidate != "" {
			motivation = candidate
		}
	}

	challenge := fallbackChallenge
	if len(challenges) > 0 {
		idx := stableIndexForDate(today, len(challenges))
		candidate := strings.TrimSpace(challenges[idx].Content)
		if candidate != "" {
			challenge = candidate
		}
	}

	return DailyContentPayload{
		Date:       dateKey,
		Motivation: motivation,
		Challenge:  challenge,
	}, nil
}

func stableIndexForDate(date time.Time, length int) int {
	if length <= 0 {
		return 0
	}
	serial := int(date.Unix() / 86400)
	if serial < 0 {
		serial = -serial
	}
	return serial % length
}

func dayStartUTC(raw time.Time) time.Time {
	value := raw.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
