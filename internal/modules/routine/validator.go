package routine

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	maxMoodLength           = 50
	maxCommitmentLength     = 2000
	maxRelapseTriggerLength = 500
	defaultWindowDays       = 30
	minWindowDays           = 7
	maxWindowDays           = 90
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
			{"field": "is_successful", "message": "Status check-in wajib diisi"},
		}, nil)
	}
	if !*req.IsSuccessful {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Gunakan endpoint relapse untuk status gagal", []map[string]string{
			{"field": "is_successful", "message": "Gunakan endpoint /api/v1/routine/relapses jika terjadi relapse"},
		}, nil)
	}

	commitment := firstNonEmptyPointer(req.Commitment, req.Content)
	if commitment != nil && len([]rune(*commitment)) > maxCommitmentLength {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Catatan check-in terlalu panjang", []map[string]string{
			{"field": "commitment", "message": "Catatan check-in maksimal 2000 karakter"},
		}, nil)
	}

	relapseTrigger, err := normalizeRelapseTriggerValues(req.RelapseTrigger)
	if err != nil {
		return DailyCheckInInput{}, err
	}
	if len(relapseTrigger) > 0 {
		return DailyCheckInInput{}, errs.New(errs.CodeValidationError, "Pemicu relapse dikirim lewat endpoint relapse", []map[string]string{
			{"field": "relapse_trigger", "message": "Gunakan endpoint /api/v1/routine/relapses untuk relapse_trigger"},
		}, nil)
	}

	return DailyCheckInInput{
		Mood:         mood,
		IsSuccessful: *req.IsSuccessful,
		JournalText:  commitment,
	}, nil
}

// NormalizeRelapseRequest validates relapse-trigger payload for explicit relapse logging.
func NormalizeRelapseRequest(req RelapseRequest) (RelapseInput, error) {
	mood := strings.TrimSpace(req.Mood)
	if mood == "" {
		return RelapseInput{}, errs.New(errs.CodeValidationError, "Mood wajib diisi", []map[string]string{
			{"field": "mood", "message": "Mood wajib diisi"},
		}, nil)
	}
	if len([]rune(mood)) > maxMoodLength {
		return RelapseInput{}, errs.New(errs.CodeValidationError, "Mood terlalu panjang", []map[string]string{
			{"field": "mood", "message": "Mood maksimal 50 karakter"},
		}, nil)
	}

	relapseTrigger, err := normalizeRelapseTriggerValues(req.RelapseTrigger)
	if err != nil {
		return RelapseInput{}, err
	}
	if len(relapseTrigger) == 0 {
		return RelapseInput{}, errs.New(errs.CodeValidationError, "Pemicu relapse wajib diisi", []map[string]string{
			{"field": "relapse_trigger", "message": "Isi minimal satu pemicu relapse"},
		}, nil)
	}

	commitment := firstNonEmptyPointer(req.Commitment, req.Content)
	if commitment != nil && len([]rune(*commitment)) > maxCommitmentLength {
		return RelapseInput{}, errs.New(errs.CodeValidationError, "Catatan relapse terlalu panjang", []map[string]string{
			{"field": "commitment", "message": "Catatan relapse maksimal 2000 karakter"},
		}, nil)
	}

	return RelapseInput{
		Mood:           mood,
		RelapseTrigger: relapseTrigger,
		JournalText:    commitment,
	}, nil
}

func normalizeRelapseTriggerValues(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	normalized := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if len([]rune(trimmed)) > maxRelapseTriggerLength {
			return nil, errs.New(errs.CodeValidationError, "Pemicu relapse terlalu panjang", []map[string]string{
				{"field": "relapse_trigger", "message": "Pemicu relapse maksimal 500 karakter"},
			}, nil)
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
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

// NormalizeActivitySummaryWindow validates and normalizes optional window days.
func NormalizeActivitySummaryWindow(raw *int) (int, error) {
	if raw == nil {
		return defaultWindowDays, nil
	}
	if *raw < minWindowDays {
		return 0, errs.New(errs.CodeValidationError, "Nilai window_days tidak valid", []map[string]string{
			{"field": "window_days", "message": "window_days minimal 7"},
		}, nil)
	}
	if *raw > maxWindowDays {
		return 0, errs.New(errs.CodeValidationError, "Nilai window_days tidak valid", []map[string]string{
			{"field": "window_days", "message": "window_days maksimal 90"},
		}, nil)
	}
	return *raw, nil
}
