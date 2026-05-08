package routine

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	maxMoodLength       = 50
	maxCommitmentLength = 2000
)

// NormalizeDailyCheckInRequest validates and normalizes check-in request payload.
func NormalizeDailyCheckInRequest(req DailyCheckInRequest) (DailyCheckInInput, error) {
	mood := strings.TrimSpace(req.Mood)
	if mood == "" {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Mood wajib diisi", []map[string]string{
			{"field": "mood", "message": "Mood wajib diisi"},
		}, nil)
	}
	if len([]rune(mood)) > maxMoodLength {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Mood terlalu panjang", []map[string]string{
			{"field": "mood", "message": "Mood maksimal 50 karakter"},
		}, nil)
	}

	if req.IsSuccessful == nil {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Status check-in wajib diisi", []map[string]string{
			{"field": "isSuccessful", "message": "Status check-in wajib diisi"},
		}, nil)
	}

	commitment := firstNonEmptyPointer(req.Commitment, req.Content)
	if commitment != nil && len([]rune(*commitment)) > maxCommitmentLength {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Catatan check-in terlalu panjang", []map[string]string{
			{"field": "commitment", "message": "Catatan check-in maksimal 2000 karakter"},
		}, nil)
	}

	return DailyCheckInInput{
		Mood:         mood,
		IsSuccessful: *req.IsSuccessful,
		JournalText:  commitment,
	}, nil
}

func firstNonEmptyPointer(candidates ...*string) *string {
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		trimmed := strings.TrimSpace(*candidate)
		if trimmed == "" {
			continue
		}
		return &trimmed
	}
	return nil
}
