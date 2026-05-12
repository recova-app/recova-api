package auth

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	timeOfDayLayout      = "15:04"
	minNicknameLength    = 3
	maxNicknameLength    = 50
	minRecoveryReasonLen = 3
	minPasswordLength    = 8
	maxPasswordBytes     = 72
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// ValidateGoogleLoginRequest validates minimal Google login request contract.
func ValidateGoogleLoginRequest(req GoogleLoginRequest) error {
	if strings.TrimSpace(req.Token) == "" {
		return errs.New(errs.CodeValidationError, "Token Google wajib diisi", []map[string]string{
			{"field": "token", "message": "Token Google wajib diisi"},
		}, nil)
	}

	return nil
}

// NormalizeAndValidateManualRegisterRequest validates and maps manual register payload.
func NormalizeAndValidateManualRegisterRequest(req ManualRegisterRequest) (ManualRegisterInput, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !isValidEmail(email) {
		return ManualRegisterInput{}, errs.New(errs.CodeValidationError, "Format email tidak valid", []map[string]string{
			{"field": "email", "message": "Format email tidak valid"},
		}, nil)
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	usernameLen := utf8.RuneCountInString(username)
	if usernameLen < minNicknameLength || usernameLen > maxNicknameLength || !usernamePattern.MatchString(username) {
		return ManualRegisterInput{}, errs.New(errs.CodeValidationError, "Username tidak valid", []map[string]string{
			{"field": "username", "message": "Username harus 3-50 karakter, hanya huruf kecil, angka, atau underscore"},
		}, nil)
	}

	password := req.Password
	if len(password) < minPasswordLength {
		return ManualRegisterInput{}, errs.New(errs.CodeValidationError, "Password minimal 8 karakter", []map[string]string{
			{"field": "password", "message": "Password minimal 8 karakter"},
		}, nil)
	}
	if len([]byte(password)) > maxPasswordBytes {
		return ManualRegisterInput{}, errs.New(errs.CodeValidationError, "Password maksimal 72 karakter", []map[string]string{
			{"field": "password", "message": "Password maksimal 72 karakter"},
		}, nil)
	}
	if password != req.ConfirmPassword {
		return ManualRegisterInput{}, errs.New(errs.CodeValidationError, "Konfirmasi password tidak cocok", []map[string]string{
			{"field": "confirm_password", "message": "Konfirmasi password tidak cocok"},
		}, nil)
	}

	return ManualRegisterInput{
		Email:    email,
		Username: username,
		Password: password,
	}, nil
}

// NormalizeAndValidateManualLoginRequest validates and maps manual login payload.
func NormalizeAndValidateManualLoginRequest(req ManualLoginRequest) (ManualLoginInput, error) {
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}
	if identifier == "" {
		return ManualLoginInput{}, errs.New(errs.CodeValidationError, "Identifier login wajib diisi", []map[string]string{
			{"field": "identifier", "message": "Identifier login wajib diisi"},
		}, nil)
	}

	password := req.Password
	if strings.TrimSpace(password) == "" {
		return ManualLoginInput{}, errs.New(errs.CodeValidationError, "Password wajib diisi", []map[string]string{
			{"field": "password", "message": "Password wajib diisi"},
		}, nil)
	}

	return ManualLoginInput{
		Identifier: strings.ToLower(identifier),
		Password:   password,
	}, nil
}

// NormalizeAndValidateOnboardingRequest validates and maps onboarding payload variants.
func NormalizeAndValidateOnboardingRequest(req OnboardingRequest) (OnboardingInput, error) {
	nickname := strings.TrimSpace(req.Nickname)
	if len([]rune(nickname)) < minNicknameLength || len([]rune(nickname)) > maxNicknameLength {
		return OnboardingInput{}, errs.New(errs.CodeValidationError, "Nama panggilan tidak valid", []map[string]string{
			{"field": "nickname", "message": "Nama panggilan harus 3-50 karakter"},
		}, nil)
	}

	recovery_reason := strings.TrimSpace(req.RecoveryReason)
	if len([]rune(recovery_reason)) < minRecoveryReasonLen {
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

	answers := req.Answers
	if answers == nil {
		answers = map[string]any{}
	}

	dependencyLevel := strings.TrimSpace(req.DependencyLevel)

	input := OnboardingInput{
		Nickname:        nickname,
		RecoveryReason:  recovery_reason,
		DailyCheckInRaw: checkInRaw,
		DailyCheckIn:    checkInTime,
		Answers:         answers,
	}
	if dependencyLevel != "" {
		input.DependencyLevel = &dependencyLevel
	}

	return input, nil
}

func isValidEmail(raw string) bool {
	if raw == "" {
		return false
	}
	parsed, err := mail.ParseAddress(raw)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(parsed.Address), raw)
}
