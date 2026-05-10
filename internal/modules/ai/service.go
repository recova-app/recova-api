package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	defaultSummaryMessage = "Insight baru untukmu akan segera tersedia. Teruslah menulis jurnal harianmu!"
	coachHistoryLimit     = 10
)

type aiRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	FindProfileByUserID(ctx context.Context, userID string) (models.Profile, error)
	FindActiveStreakByUserID(ctx context.Context, userID string) (*models.Streak, error)
	ListRecentChatsByUserID(ctx context.Context, userID string, limit int) ([]models.AIChat, error)
	CreateChatMessages(ctx context.Context, rows []models.AIChat) error
	GetPersonaPreferenceByUserID(ctx context.Context, userID string) (models.UserAIPersonaPreference, error)
	UpsertPersonaPreference(ctx context.Context, userID string, persona string, updatedAt time.Time) error
}

type aiTelemetry interface {
	RecordPersonaUsage(action string, persona string, err error)
}

type noopTelemetry struct{}

func (noopTelemetry) RecordPersonaUsage(_ string, _ string, _ error) {}

const (
	telemetryActionAskCoach          = "ask_coach"
	telemetryActionPersonaGet        = "persona_preference_get"
	telemetryActionPersonaPreference = "persona_preference_update"
)

// Service owns AI coach business rules and orchestration.
type Service struct {
	repo      aiRepository
	provider  aiplatform.Client
	telemetry aiTelemetry
	now       func() time.Time
}

// NewService constructs AI service.
func NewService(repo aiRepository, provider aiplatform.Client) *Service {
	return &Service{
		repo:      repo,
		provider:  provider,
		telemetry: noopTelemetry{},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// SetTelemetry configures optional persona telemetry recorder.
func (s *Service) SetTelemetry(telemetry aiTelemetry) {
	if s == nil || telemetry == nil {
		return
	}
	s.telemetry = telemetry
}

// AskCoach sends one user message to AI provider, stores conversation, and returns safe reply.
func (s *Service) AskCoach(ctx context.Context, userID string, req AskCoachRequest) (AskCoachResponseData, error) {
	input, err := NormalizeAskCoachRequest(req)
	if err != nil {
		return AskCoachResponseData{}, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return AskCoachResponseData{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return AskCoachResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	profile, profileErr := s.repo.FindProfileByUserID(ctx, userID)
	if profileErr != nil && !IsRecordNotFound(profileErr) {
		return AskCoachResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data profil", nil, profileErr)
	}

	streakDays := 0
	activeStreak, err := s.repo.FindActiveStreakByUserID(ctx, userID)
	if err != nil {
		return AskCoachResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data streak", nil, err)
	}
	if activeStreak != nil && !activeStreak.StartDate.IsZero() {
		streakDays = calculateStreakDays(s.now(), activeStreak.StartDate)
	}

	history, err := s.repo.ListRecentChatsByUserID(ctx, userID, coachHistoryLimit)
	if err != nil {
		return AskCoachResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca riwayat chat AI", nil, err)
	}

	persona, err := s.resolvePersonaPreference(ctx, userID)
	if err != nil {
		return AskCoachResponseData{}, err
	}

	reply, err := s.provider.Generate(ctx, aiplatform.GenerateRequest{
		SystemInstruction: buildCoachSystemInstruction(user, profile, streakDays, persona),
		UserPrompt:        input.Message,
		History:           mapAIHistory(history),
		Persona:           persona,
	})
	if err != nil {
		s.telemetry.RecordPersonaUsage(telemetryActionAskCoach, persona, err)
		return AskCoachResponseData{}, mapProviderError(err)
	}

	chatRows := []models.AIChat{
		{UserID: strings.TrimSpace(userID), Role: "user", Content: input.Message},
		{UserID: strings.TrimSpace(userID), Role: "model", Content: strings.TrimSpace(reply.Text)},
	}
	if err := s.repo.CreateChatMessages(ctx, chatRows); err != nil {
		s.telemetry.RecordPersonaUsage(telemetryActionAskCoach, persona, err)
		return AskCoachResponseData{}, errs.New(errs.CodeInternalError, "Gagal menyimpan riwayat chat AI", nil, err)
	}

	s.telemetry.RecordPersonaUsage(telemetryActionAskCoach, persona, nil)
	return AskCoachResponseData{
		Response:    strings.TrimSpace(reply.Text),
		PersonaUsed: persona,
	}, nil
}

// GetChatHistory returns authenticated user chat history.
func (s *Service) GetChatHistory(ctx context.Context, userID string, query ChatHistoryQuery) ([]ChatHistoryItem, error) {
	limit, err := NormalizeChatHistoryLimit(query.Limit)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return nil, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	rows, err := s.repo.ListRecentChatsByUserID(ctx, userID, limit)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca riwayat chat AI", nil, err)
	}

	payload := make([]ChatHistoryItem, 0, len(rows))
	for _, row := range rows {
		payload = append(payload, ChatHistoryItem{
			ID:        row.ID,
			Role:      strings.TrimSpace(row.Role),
			Content:   row.Content,
			CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return payload, nil
}

// GetSummary returns latest AI summary bound to authenticated user.
func (s *Service) GetSummary(ctx context.Context, userID string) (SummaryResponseData, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return SummaryResponseData{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return SummaryResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	profile, err := s.repo.FindProfileByUserID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return SummaryResponseData{Summary: defaultSummaryMessage}, nil
		}
		return SummaryResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca ringkasan pengguna", nil, err)
	}

	summary := strings.TrimSpace(valueOrEmpty(profile.AISummary))
	if summary == "" {
		summary = defaultSummaryMessage
	}
	return SummaryResponseData{Summary: summary}, nil
}

// AnalyzeOnboarding calls AI provider to classify onboarding answers into safe structured output.
func (s *Service) AnalyzeOnboarding(ctx context.Context, userID string, req OnboardingAnalysisRequest) (OnboardingAnalysisResponseData, error) {
	input, err := NormalizeOnboardingAnalysisRequest(req)
	if err != nil {
		return OnboardingAnalysisResponseData{}, err
	}

	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return OnboardingAnalysisResponseData{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return OnboardingAnalysisResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	generated, err := s.provider.Generate(ctx, aiplatform.GenerateRequest{
		SystemInstruction: onboardingSystemInstruction,
		UserPrompt:        buildOnboardingPrompt(input.Answers),
		ForceJSON:         true,
	})
	if err != nil {
		return OnboardingAnalysisResponseData{}, mapProviderError(err)
	}

	result, err := parseOnboardingAnalysisJSON(generated.Text)
	if err != nil {
		return OnboardingAnalysisResponseData{}, errs.New(errs.CodeDownstreamError, "Respons analisis onboarding tidak valid", nil, err)
	}
	return result, nil
}

// GetPersonaPreference returns user persona preference with safe fallback default.
func (s *Service) GetPersonaPreference(ctx context.Context, userID string) (PersonaPreferenceResponseData, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return PersonaPreferenceResponseData{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return PersonaPreferenceResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	persona := DefaultPersona
	row, err := s.repo.GetPersonaPreferenceByUserID(ctx, userID)
	if err != nil {
		if !IsRecordNotFound(err) {
			return PersonaPreferenceResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca preferensi persona AI", nil, err)
		}
	} else {
		persona = ResolvePersonaOrDefault(row.Persona)
	}

	s.telemetry.RecordPersonaUsage(telemetryActionPersonaGet, persona, nil)
	return PersonaPreferenceResponseData{
		Persona:         persona,
		FallbackPersona: DefaultPersona,
	}, nil
}

// UpdatePersonaPreference validates and persists user persona preference.
func (s *Service) UpdatePersonaPreference(ctx context.Context, userID string, req PersonaPreferenceRequest) (PersonaPreferenceResponseData, error) {
	input, err := NormalizePersonaPreferenceRequest(req)
	if err != nil {
		return PersonaPreferenceResponseData{}, err
	}

	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return PersonaPreferenceResponseData{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return PersonaPreferenceResponseData{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	if err := s.repo.UpsertPersonaPreference(ctx, userID, input.Persona, s.now()); err != nil {
		s.telemetry.RecordPersonaUsage(telemetryActionPersonaPreference, input.Persona, err)
		return PersonaPreferenceResponseData{}, errs.New(errs.CodeInternalError, "Gagal menyimpan preferensi persona AI", nil, err)
	}

	s.telemetry.RecordPersonaUsage(telemetryActionPersonaPreference, input.Persona, nil)
	return PersonaPreferenceResponseData{
		Persona:         input.Persona,
		FallbackPersona: DefaultPersona,
	}, nil
}

func mapProviderError(err error) error {
	switch aiplatform.ClassifyError(err) {
	case aiplatform.ErrorKindTimeout, aiplatform.ErrorKindUnavailable:
		return errs.New(errs.CodeServiceUnavailable, "Layanan AI sedang tidak tersedia, coba lagi", nil, err)
	case aiplatform.ErrorKindInvalidResponse, aiplatform.ErrorKindUnauthorized:
		return errs.New(errs.CodeDownstreamError, "Layanan AI sedang bermasalah", nil, err)
	default:
		return errs.New(errs.CodeDownstreamError, "Terjadi kegagalan pada layanan AI", nil, err)
	}
}

func mapAIHistory(rows []models.AIChat) []aiplatform.Message {
	result := make([]aiplatform.Message, 0, len(rows))
	for _, row := range rows {
		content := strings.TrimSpace(row.Content)
		if content == "" {
			continue
		}
		result = append(result, aiplatform.Message{
			Role:    strings.TrimSpace(row.Role),
			Content: content,
		})
	}
	return result
}

func calculateStreakDays(nowUTC time.Time, startDate time.Time) int {
	start := startDate.UTC()
	now := nowUTC.UTC()
	if start.After(now) {
		return 0
	}
	days := int(now.Sub(start).Hours()/24) + 1
	if days < 0 {
		return 0
	}
	return days
}

func buildCoachSystemInstruction(user models.User, profile models.Profile, streakDays int, persona string) string {
	nickname := strings.TrimSpace(user.Nickname)
	if nickname == "" {
		nickname = "Friend"
	}
	why := strings.TrimSpace(valueOrEmpty(user.UserWhy))
	if why == "" {
		why = "maintaining recovery commitment"
	}

	dependencyLevel := strings.TrimSpace(valueOrEmpty(profile.DependencyLevel))
	if dependencyLevel == "" {
		dependencyLevel = "unknown"
	}

	personaActive := ResolvePersonaOrDefault(persona)

	return fmt.Sprintf(`You are Recova AI Coach: empathetic, supportive, and non-judgmental. Respond in English.
Keep the conversation focused on recovery support from pornography addiction.
If the user asks about out-of-scope topics, politely decline and redirect to recovery.
Use concise 1-3 sentence responses, warm tone, and include one small actionable step.
Do not provide medical diagnosis or shame the user.
Active persona:
- persona name: %s
- style guidance: %s
User context:
- nickname: %s
- streak days: %d
- recovery reason: %s
- onboarding dependency level: %s`, personaActive, personaStyleInstruction(personaActive), nickname, streakDays, why, dependencyLevel)
}

const onboardingSystemInstruction = `You are a Recova onboarding analyst. Always respond with ONLY valid JSON, without markdown.
Required schema:
{"level":"Low|Moderate|High","title":"...","level_description":"...","pattern_analysis":"...","encouragement":"..."}
Use English with a supportive, non-shaming tone.`

func buildOnboardingPrompt(answers map[string]any) string {
	keys := make([]string, 0, len(answers))
	for key := range answers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("Briefly analyze the following onboarding answers:\n")
	for _, key := range keys {
		value := answers[key]
		builder.WriteString("- ")
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(formatAnswerValue(value))
		builder.WriteString("\n")
	}
	builder.WriteString("Classify dependency level and provide realistic encouragement.")
	return builder.String()
}

func formatAnswerValue(value any) string {
	if value == nil {
		return "null"
	}
	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return ""
		}
		return trimmed
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(raw)
	}
}

func parseOnboardingAnalysisJSON(raw string) (OnboardingAnalysisResponseData, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return OnboardingAnalysisResponseData{}, fmt.Errorf("empty analysis response")
	}

	var parsed struct {
		Level            string `json:"level"`
		Title            string `json:"title"`
		LevelDescription string `json:"level_description"`
		PatternAnalysis  string `json:"pattern_analysis"`
		Encouragement    string `json:"encouragement"`
	}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return OnboardingAnalysisResponseData{}, err
	}

	result := OnboardingAnalysisResponseData{
		Level:            strings.TrimSpace(parsed.Level),
		Title:            strings.TrimSpace(parsed.Title),
		LevelDescription: strings.TrimSpace(parsed.LevelDescription),
		PatternAnalysis:  strings.TrimSpace(parsed.PatternAnalysis),
		Encouragement:    strings.TrimSpace(parsed.Encouragement),
	}

	if result.Level == "" || result.Title == "" || result.LevelDescription == "" || result.PatternAnalysis == "" || result.Encouragement == "" {
		return OnboardingAnalysisResponseData{}, fmt.Errorf("incomplete analysis response")
	}

	return result, nil
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (s *Service) resolvePersonaPreference(ctx context.Context, userID string) (string, error) {
	row, err := s.repo.GetPersonaPreferenceByUserID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return DefaultPersona, nil
		}
		return "", errs.New(errs.CodeInternalError, "Gagal membaca preferensi persona AI", nil, err)
	}
	return ResolvePersonaOrDefault(row.Persona), nil
}

func personaStyleInstruction(persona string) string {
	switch ResolvePersonaOrDefault(persona) {
	case "friendly":
		return "use friendly, warm, and light language while preserving safety boundaries"
	case "concise":
		return "focus on the core point, keep answers concise, stay empathetic, avoid long sentences"
	case "direct":
		return "be firm and to-the-point, provide clear action steps without judgment"
	default:
		return "supportive, empathetic, and calming"
	}
}
