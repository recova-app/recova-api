package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	timeOfDayLayout      = "15:04"
	minNicknameLength    = 3
	maxNicknameLength    = 50
	minRecoveryReasonLen = 3
)

// ValidateGoogleLoginRequest validates minimal Google login request contract.
func ValidateGoogleLoginRequest(req GoogleLoginRequest) error {
	if strings.TrimSpace(req.Token) == "" {
		return errs.New(errs.CodeValidationError, "Token Google wajib diisi", []map[string]string{
			{"field": "token", "message": "Token Google wajib diisi"},
		}, nil)
	}

	return nil
}

// NormalizeAndValidateOnboardingRequest validates and maps onboarding payload variants.
func NormalizeAndValidateOnboardingRequest(req OnboardingRequest) (OnboardingInput, error) {
	nickname := strings.TrimSpace(req.Nickname)
	if len([]rune(nickname)) < minNicknameLength || len([]rune(nickname)) > maxNicknameLength {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Nama panggilan tidak valid", []map[string]string{
			{"field": "nickname", "message": "Nama panggilan harus 3-50 karakter"},
		}, nil)
	}

	recoveryReason := strings.TrimSpace(req.RecoveryReason)
	if recoveryReason == "" {
		recoveryReason = strings.TrimSpace(req.RecoveryReasonLegacy)
	}
	if len([]rune(recoveryReason)) < minRecoveryReasonLen {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Alasan pemulihan wajib diisi", []map[string]string{
			{"field": "recovery_reason", "message": "Alasan pemulihan minimal 3 karakter"},
		}, nil)
	}

	checkInRaw := strings.TrimSpace(req.DailyCheckInTime)
	if checkInRaw == "" {
		checkInRaw = strings.TrimSpace(req.DailyCheckInTimeLegacy)
	}
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

	answers := req.Answers
	if answers == nil {
		answers = map[string]any{}
	}

	dependencyLevel := strings.TrimSpace(req.DependencyLevel)
	if dependencyLevel == "" {
		dependencyLevel = strings.TrimSpace(req.DependencyLevelLegacy)
	}

	input := OnboardingInput{
		Nickname:        nickname,
		RecoveryReason:  recoveryReason,
		DailyCheckInRaw: checkInRaw,
		DailyCheckIn:    checkInTime,
		Answers:         answers,
	}
	if dependencyLevel != "" {
		input.DependencyLevel = &dependencyLevel
	}

	return input, nil
}
