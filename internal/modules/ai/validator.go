package ai

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	defaultHistoryLimit  = 50
	maxHistoryLimit      = 200
	maxPromptLength      = 4000
	maxMoodLength        = 50
	maxRelapseTextLength = 500

	// DefaultPersona is safe fallback persona when user preference is empty/invalid.
	DefaultPersona = "supportive"
)

var allowedPersonaSet = map[string]struct{}{
	DefaultPersona: {},
	"friendly":     {},
	"concise":      {},
	"direct":       {},
}

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

// NormalizePersonaPreferenceRequest validates and normalizes persona preference payload.
func NormalizePersonaPreferenceRequest(req PersonaPreferenceRequest) (PersonaPreferenceInput, error) {
	persona, ok := NormalizePersona(req.Persona)
	if !ok {
		return PersonaPreferenceInput{}, errs.New(errs.CodeValidationError, "Nilai persona tidak valid", []map[string]string{{
			"field": "persona", "message": "Persona harus salah satu dari supportive, friendly, concise, direct",
		}}, nil)
	}
	return PersonaPreferenceInput{Persona: persona}, nil
}

// NormalizeRelapseSolutionRequest validates and normalizes relapse-solution payload.
func NormalizeRelapseSolutionRequest(req RelapseSolutionRequest) (RelapseSolutionInput, error) {
	mood := strings.TrimSpace(req.Mood)
	if mood == "" {
		return RelapseSolutionInput{}, errs.New(errs.CodeValidationError, "Mood wajib diisi", []map[string]string{{
			"field": "mood", "message": "Mood wajib diisi",
		}}, nil)
	}
	if len([]rune(mood)) > maxMoodLength {
		return RelapseSolutionInput{}, errs.New(errs.CodeValidationError, "Mood terlalu panjang", []map[string]string{{
			"field": "mood", "message": "Mood maksimal 50 karakter",
		}}, nil)
	}

	relapseTrigger, err := normalizeRelapseTriggerValues(req.RelapseTrigger)
	if err != nil {
		return RelapseSolutionInput{}, err
	}

	commitment := normalizeOptionalText(req.Commitment)
	if commitment != nil && len([]rune(*commitment)) > maxPromptLength {
		return RelapseSolutionInput{}, errs.New(errs.CodeValidationError, "Catatan terlalu panjang", []map[string]string{{
			"field": "commitment", "message": "Catatan maksimal 4000 karakter",
		}}, nil)
	}

	return RelapseSolutionInput{
		Mood:           mood,
		RelapseTrigger: relapseTrigger,
		Commitment:     commitment,
	}, nil
}

// NormalizePersona validates persona enum and returns normalized lower-case value.
func NormalizePersona(raw string) (string, bool) {
	persona := strings.ToLower(strings.TrimSpace(raw))
	if persona == "" {
		return "", false
	}
	if _, ok := allowedPersonaSet[persona]; !ok {
		return "", false
	}
	return persona, true
}

// ResolvePersonaOrDefault returns safe default persona if value is empty/invalid.
func ResolvePersonaOrDefault(raw string) string {
	if persona, ok := NormalizePersona(raw); ok {
		return persona
	}
	return DefaultPersona
}

func normalizeOptionalText(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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
		if len([]rune(trimmed)) > maxRelapseTextLength {
			return nil, errs.New(errs.CodeValidationError, "Pemicu relapse terlalu panjang", []map[string]string{{
				"field": "relapse_trigger", "message": "Pemicu relapse maksimal 500 karakter",
			}}, nil)
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}
