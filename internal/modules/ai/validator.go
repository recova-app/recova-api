package ai

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	defaultHistoryLimit = 50
	maxHistoryLimit     = 200
	maxPromptLength     = 4000
)

// NormalizeAskCoachRequest validates and normalizes ask-coach payload.
func NormalizeAskCoachRequest(req AskCoachRequest) (AskCoachInput, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return AskCoachInput{}, errs.New(errs.CodeValidationError, "Pesan wajib diisi", []map[string]string{{
			"field": "message", "message": "Pesan wajib diisi",
		}}, nil)
	}
	if len([]rune(message)) > maxPromptLength {
		return AskCoachInput{}, errs.New(errs.CodeValidationError, "Pesan terlalu panjang", []map[string]string{{
			"field": "message", "message": "Pesan maksimal 4000 karakter",
		}}, nil)
	}

	return AskCoachInput{Message: message}, nil
}

// NormalizeOnboardingAnalysisRequest validates and normalizes onboarding-analysis payload.
func NormalizeOnboardingAnalysisRequest(req OnboardingAnalysisRequest) (OnboardingAnalysisInput, error) {
	if len(req.Answers) == 0 {
		return OnboardingAnalysisInput{}, errs.New(errs.CodeValidationError, "Jawaban onboarding wajib diisi", []map[string]string{{
			"field": "answers", "message": "Jawaban onboarding wajib diisi",
		}}, nil)
	}

	answers := make(map[string]any, len(req.Answers))
	for key, value := range req.Answers {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		answers[trimmed] = value
	}
	if len(answers) == 0 {
		return OnboardingAnalysisInput{}, errs.New(errs.CodeValidationError, "Jawaban onboarding wajib diisi", []map[string]string{{
			"field": "answers", "message": "Jawaban onboarding wajib diisi",
		}}, nil)
	}

	return OnboardingAnalysisInput{Answers: answers}, nil
}

// NormalizeChatHistoryLimit validates and normalizes optional chat-history limit.
func NormalizeChatHistoryLimit(raw *int) (int, error) {
	if raw == nil {
		return defaultHistoryLimit, nil
	}
	if *raw < 1 {
		return 0, errs.New(errs.CodeValidationError, "Nilai limit tidak valid", []map[string]string{{
			"field": "limit", "message": "Limit minimal 1",
		}}, nil)
	}
	if *raw > maxHistoryLimit {
		return 0, errs.New(errs.CodeValidationError, "Nilai limit tidak valid", []map[string]string{{
			"field": "limit", "message": "Limit maksimal 200",
		}}, nil)
	}
	return *raw, nil
}
