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
}

// Service owns AI coach business rules and orchestration.
type Service struct {
	repo     aiRepository
	provider aiplatform.Client
	now      func() time.Time
}

// NewService constructs AI service.
func NewService(repo aiRepository, provider aiplatform.Client) *Service {
	return &Service{
		repo:     repo,
		provider: provider,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
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

	reply, err := s.provider.Generate(ctx, aiplatform.GenerateRequest{
		SystemInstruction: buildCoachSystemInstruction(user, profile, streakDays),
		UserPrompt:        input.Message,
		History:           mapAIHistory(history),
	})
	if err != nil {
		return AskCoachResponseData{}, mapProviderError(err)
	}

	chatRows := []models.AIChat{
		{UserID: strings.TrimSpace(userID), Role: "user", Content: input.Message},
		{UserID: strings.TrimSpace(userID), Role: "model", Content: strings.TrimSpace(reply.Text)},
	}
	if err := s.repo.CreateChatMessages(ctx, chatRows); err != nil {
		return AskCoachResponseData{}, errs.New(errs.CodeInternalError, "Gagal menyimpan riwayat chat AI", nil, err)
	}

	return AskCoachResponseData{Response: strings.TrimSpace(reply.Text)}, nil
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

func buildCoachSystemInstruction(user models.User, profile models.Profile, streakDays int) string {
	nickname := strings.TrimSpace(user.Nickname)
	if nickname == "" {
		nickname = "Teman"
	}
	why := strings.TrimSpace(valueOrEmpty(user.UserWhy))
	if why == "" {
		why = "menjaga komitmen pemulihan"
	}

	dependencyLevel := strings.TrimSpace(valueOrEmpty(profile.DependencyLevel))
	if dependencyLevel == "" {
		dependencyLevel = "belum diketahui"
	}

	return fmt.Sprintf(`Kamu adalah Recova AI Coach yang empatik, suportif, dan tidak menghakimi. Gunakan Bahasa Indonesia.
Fokus percakapan hanya untuk dukungan pemulihan dari kecanduan pornografi.
Jika user bertanya topik di luar konteks, tolak dengan sopan lalu arahkan kembali ke topik pemulihan.
Gunakan jawaban singkat 1-3 kalimat, hangat, dan berikan satu langkah kecil yang dapat dilakukan sekarang.
Jangan memberi diagnosis medis atau menyalahkan user.
Konteks user:
- nickname: %s
- streak hari: %d
- alasan pemulihan: %s
- level ketergantungan onboarding: %s`, nickname, streakDays, why, dependencyLevel)
}

const onboardingSystemInstruction = `Kamu adalah analis onboarding Recova. Selalu jawab HANYA JSON valid tanpa markdown.
Skema wajib:
{"level":"Rendah|Sedang|Tinggi","title":"...","level_description":"...","pattern_analysis":"...","encouragement":"..."}
Gunakan Bahasa Indonesia, nada suportif, tanpa menyalahkan.`

func buildOnboardingPrompt(answers map[string]any) string {
	keys := make([]string, 0, len(answers))
	for key := range answers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("Analisis jawaban onboarding berikut secara ringkas:\n")
	for _, key := range keys {
		value := answers[key]
		builder.WriteString("- ")
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(formatAnswerValue(value))
		builder.WriteString("\n")
	}
	builder.WriteString("Klasifikasikan level ketergantungan dan berikan encouragement yang realistis.")
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
		Level                 string `json:"level"`
		Title                 string `json:"title"`
		LevelDescription      string `json:"levelDescription"`
		LevelDescriptionSnake string `json:"level_description"`
		PatternAnalysis       string `json:"patternAnalysis"`
		PatternAnalysisSnake  string `json:"pattern_analysis"`
		Encouragement         string `json:"encouragement"`
	}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return OnboardingAnalysisResponseData{}, err
	}

	levelDescription := strings.TrimSpace(parsed.LevelDescription)
	if levelDescription == "" {
		levelDescription = strings.TrimSpace(parsed.LevelDescriptionSnake)
	}
	patternAnalysis := strings.TrimSpace(parsed.PatternAnalysis)
	if patternAnalysis == "" {
		patternAnalysis = strings.TrimSpace(parsed.PatternAnalysisSnake)
	}

	result := OnboardingAnalysisResponseData{
		Level:            strings.TrimSpace(parsed.Level),
		Title:            strings.TrimSpace(parsed.Title),
		LevelDescription: levelDescription,
		PatternAnalysis:  patternAnalysis,
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
