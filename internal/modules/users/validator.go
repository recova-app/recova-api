package users

import (
	"fmt"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	timeOfDayLayout         = "15:04"
	minNicknameLength       = 3
	maxNicknameLength       = 50
	minRecoveryReasonLength = 3
	minPornFreeGoalDays     = 1
	maxPornFreeGoalDays     = 3650
)

// NormalizeSettingsUpdate validates update payload and returns sanitized updates.
func NormalizeSettingsUpdate(req SettingsUpdateRequest) (map[string]any, error) {
	updates := map[string]any{}

	if req.Nickname != nil {
		nickname := strings.TrimSpace(*req.Nickname)
		if len([]rune(nickname)) < minNicknameLength || len([]rune(nickname)) > maxNicknameLength {
			return nil, errs.New(errs.CodeValidationError, "Nama panggilan tidak valid", []map[string]string{
				{"field": "nickname", "message": "Nama panggilan harus 3-50 karakter"},
			}, nil)
		}
		updates["nickname"] = nickname
	}

	recovery_reason := firstNonEmptyPointer(req.RecoveryReason)
	if recovery_reason != nil {
		normalized := strings.TrimSpace(*recovery_reason)
		if len([]rune(normalized)) < minRecoveryReasonLength {
			return nil, errs.New(errs.CodeValidationError, "Alasan pemulihan tidak valid", []map[string]string{
				{"field": "recovery_reason", "message": "Alasan pemulihan minimal 3 karakter"},
			}, nil)
		}
		updates["user_why"] = normalized
	}

	timeRaw := firstNonEmptyPointer(req.DailyCheckInTime)
	if timeRaw != nil {
		parsed, err := time.Parse(timeOfDayLayout, strings.TrimSpace(*timeRaw))
		if err != nil {
			return nil, errs.New(errs.CodeValidationError, "Format waktu check-in tidak valid", []map[string]string{
				{"field": "daily_checkin_time", "message": "Gunakan format HH:mm"},
			}, fmt.Errorf("parse checkin time: %w", err))
		}
		updates["check_in_time"] = parsed
	}

	if len(updates) == 0 {
		return nil, errs.New(errs.CodeValidationError, "Payload update kosong", []map[string]string{
			{"field": "body", "message": "Setidaknya satu field pengaturan wajib diisi"},
		}, nil)
	}

	return updates, nil
}

// NormalizeOnboardingRequest validates onboarding payload and returns sanitized input.
func NormalizeOnboardingRequest(req OnboardingRequest) (OnboardingInput, error) {
	nickname := strings.TrimSpace(req.Nickname)
	if len([]rune(nickname)) < minNicknameLength || len([]rune(nickname)) > maxNicknameLength {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Nama panggilan tidak valid", []map[string]string{
			{"field": "nickname", "message": "Nama panggilan harus 3-50 karakter"},
		}, nil)
	}

	recovery_reason := strings.TrimSpace(req.RecoveryReason)
	if len([]rune(recovery_reason)) < minRecoveryReasonLength {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Alasan pemulihan wajib diisi", []map[string]string{
			{"field": "recovery_reason", "message": "Alasan pemulihan minimal 3 karakter"},
		}, nil)
	}

	checkInRaw := strings.TrimSpace(req.DailyCheckInTime)
	if checkInRaw == "" {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Waktu check-in harian wajib diisi", []map[string]string{
			{"field": "daily_checkin_time", "message": "Waktu check-in harian wajib diisi"},
		}, nil)
	}

	checkInTime, err := time.Parse(timeOfDayLayout, checkInRaw)
	if err != nil {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Format waktu check-in tidak valid", []map[string]string{
			{"field": "daily_checkin_time", "message": "Gunakan format HH:mm"},
		}, fmt.Errorf("parse daily check-in time: %w", err))
	}

	if req.PornFreeGoal == nil {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Target bebas pornografi wajib diisi", []map[string]string{
			{"field": "porn_free_goal", "message": "Target bebas pornografi wajib diisi"},
		}, nil)
	}

	pornFreeGoal := *req.PornFreeGoal
	if pornFreeGoal < minPornFreeGoalDays || pornFreeGoal > maxPornFreeGoalDays {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Target bebas pornografi tidak valid", []map[string]string{
			{"field": "porn_free_goal", "message": "Target bebas pornografi harus 1-3650 hari"},
		}, nil)
	}

	answers := req.Answers
	if answers == nil {
		answers = map[string]any{}
	}

	dependency := strings.TrimSpace(req.DependencyLevel)

	input := OnboardingInput{
		Nickname:        nickname,
		RecoveryReason:  recovery_reason,
		DailyCheckInRaw: checkInRaw,
		DailyCheckIn:    checkInTime,
		PornFreeGoal:    pornFreeGoal,
		Answers:         answers,
	}
	if dependency != "" {
		input.DependencyLevel = &dependency
	}

	return input, nil
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
