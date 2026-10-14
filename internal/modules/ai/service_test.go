package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

func TestService_AskCoach_ProviderTimeoutMapsServiceUnavailable(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	provider := &fakeAIProvider{err: &aiplatform.ProviderError{Provider: aiplatform.ProviderGemini, Kind: aiplatform.ErrorKindTimeout, Message: "timeout"}}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "Hello"})
	if err == nil {
		t.Fatal("expected error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeServiceUnavailable {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %s", mapped.Code)
	}
}

func TestService_AskCoach_ProviderErrorMessageIsSafe(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	provider := &fakeAIProvider{err: &aiplatform.ProviderError{
		Provider: aiplatform.ProviderOpenAICompatible,
		Kind:     aiplatform.ErrorKindInvalidResponse,
		Message:  "payload rejected: user prompt 'highly sensitive'",
	}}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "highly sensitive"})
	if err == nil {
		t.Fatal("expected error")
	}

	mapped := errs.Map(err)
	if mapped.Code != errs.CodeDownstreamError {
		t.Fatalf("expected DOWNSTREAM_ERROR, got %s", mapped.Code)
	}
	if strings.Contains(strings.ToLower(mapped.Message), "highly sensitive") {
		t.Fatalf("expected safe message, got %q", mapped.Message)
	}
}

func TestService_AskCoach_SuccessStoresConversation(t *testing.T) {
	profileReason := "want to be healthy"
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test", UserWhy: &profileReason},
		history: []models.AIChat{{Role: "user", Content: "old message"}},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Stay strong"}}
	service := NewService(repo, provider)
	service.now = func() time.Time {
		return time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC)
	}

	payload, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "Today is hard"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Response != "Stay strong" {
		t.Fatalf("unexpected response: %q", payload.Response)
	}
	if payload.PersonaUsed != DefaultPersona {
		t.Fatalf("expected default persona %q, got %q", DefaultPersona, payload.PersonaUsed)
	}
	if len(repo.createdMessages) != 2 {
		t.Fatalf("expected 2 stored rows, got %d", len(repo.createdMessages))
	}
	if repo.createdMessages[0].Role != "user" || repo.createdMessages[1].Role != "assistant" {
		t.Fatalf("unexpected stored roles: %+v", repo.createdMessages)
	}
}

func TestService_AskCoach_StripsLeakedPersonaMarkerFromReply(t *testing.T) {
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		persona: models.UserAIPersonaPreference{UserID: "user-1", Persona: "direct"},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Ok.\n\nrecova.persona.direct.v1"}}
	service := NewService(repo, provider)

	payload, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Response != "Ok." {
		t.Fatalf("expected sanitized reply, got %q", payload.Response)
	}
	if len(repo.createdMessages) != 2 {
		t.Fatalf("expected 2 stored rows, got %d", len(repo.createdMessages))
	}
	if repo.createdMessages[1].Content != "Ok." {
		t.Fatalf("expected sanitized stored assistant message, got %q", repo.createdMessages[1].Content)
	}
}

func TestService_GetChatHistory_NormalizesRoleToContract(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		history: []models.AIChat{
			{ID: "chat-1", Role: "model", Content: "A", CreatedAt: time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC)},
			{ID: "chat-2", Role: "user", Content: "B", CreatedAt: time.Date(2026, 5, 8, 9, 1, 0, 0, time.UTC)},
		},
	}
	service := NewService(repo, &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "ok"}})

	payload, err := service.GetChatHistory(context.Background(), "user-1", ChatHistoryQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected 2 history rows, got %d", len(payload))
	}
	if payload[0].Role != "assistant" {
		t.Fatalf("expected model role normalized to assistant, got %q", payload[0].Role)
	}
	if payload[1].Role != "user" {
		t.Fatalf("expected user role kept as user, got %q", payload[1].Role)
	}
}

func TestService_GetSummary_FallbackWhenNoProfile(t *testing.T) {
	repo := &fakeAIRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		profileErr: gorm.ErrRecordNotFound,
	}
	service := NewService(repo, &fakeAIProvider{})

	payload, err := service.GetSummary(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Summary != defaultSummaryMessage {
		t.Fatalf("expected fallback summary, got %q", payload.Summary)
	}
}

func TestService_AnalyzeOnboarding_InvalidProviderJSON(t *testing.T) {
	repo := &fakeAIRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "not-json"}}
	service := NewService(repo, provider)

	_, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frequency": "daily"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeDownstreamError {
		t.Fatalf("expected DOWNSTREAM_ERROR, got %s", mapped.Code)
	}
}

func TestService_AnalyzeOnboarding_Success(t *testing.T) {
	repo := &fakeAIRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: `{"level":"Moderate","title":"Initial Analysis","level_description":"Explanation","pattern_analysis":"Stress pattern","encouragement":"You can do this"}`}}
	service := NewService(repo, provider)

	payload, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frequency": "daily"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Level != "Moderate" || payload.PatternAnalysis != "Stress pattern" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestService_AnalyzeOnboarding_InvalidLevel(t *testing.T) {
	repo := &fakeAIRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: `{"level":"Unknown","title":"Initial Analysis","level_description":"Explanation","pattern_analysis":"Stress pattern","encouragement":"You can do this"}`}}
	service := NewService(repo, provider)

	_, err := service.AnalyzeOnboarding(context.Background(), "user-1", OnboardingAnalysisRequest{
		Answers: map[string]any{"frequency": "daily"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeDownstreamError {
		t.Fatalf("expected DOWNSTREAM_ERROR, got %s", mapped.Code)
	}
}

func TestService_GenerateRelapseSolution_Success(t *testing.T) {
	repo := &fakeAIRepo{user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"}}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{
		Text: `{"title":"Stabilkan Diri","analysis":"Pemicu utama terlihat dari kebiasaan scrolling saat lelah.","summary":"Solusi terbaik: blok akses perangkat pemicu di jam rawan lalu pindah ke aktivitas fisik singkat."}`,
	}}
	service := NewService(repo, provider)

	payload, err := service.GenerateRelapseSolution(context.Background(), "user-1", RelapseSolutionRequest{
		Mood:           "cemas",
		RelapseTrigger: []string{"scrolling malam", "sendiri kamar"},
		Commitment:     ptrString("ingin kembali fokus"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Title == "" || payload.Analysis == "" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Summary == "" {
		t.Fatalf("expected relapse summary, got %+v", payload)
	}
}

func TestService_GetPersonaPreference_FallbackDefault(t *testing.T) {
	repo := &fakeAIRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		personaErr: gorm.ErrRecordNotFound,
	}
	service := NewService(repo, &fakeAIProvider{})

	payload, err := service.GetPersonaPreference(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Persona != DefaultPersona {
		t.Fatalf("expected default persona %q, got %q", DefaultPersona, payload.Persona)
	}
	if payload.FallbackPersona != DefaultPersona {
		t.Fatalf("unexpected fallback persona: %q", payload.FallbackPersona)
	}
}

func TestService_UpdatePersonaPreference_Validation(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	service := NewService(repo, &fakeAIProvider{})

	_, err := service.UpdatePersonaPreference(context.Background(), "user-1", PersonaPreferenceRequest{Persona: "invalid"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if mapped := errs.Map(err); mapped.Code != errs.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR, got %s", mapped.Code)
	}
}

func TestService_AskCoach_UsesStoredPersonaInProviderRequest(t *testing.T) {
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		persona: models.UserAIPersonaPreference{UserID: "user-1", Persona: "friendly"},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "ok"}}
	service := NewService(repo, provider)

	payload, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.PersonaUsed != "friendly" {
		t.Fatalf("expected persona_used friendly, got %q", payload.PersonaUsed)
	}
	if provider.lastReq.Persona != "friendly" {
		t.Fatalf("expected provider persona friendly, got %q", provider.lastReq.Persona)
	}
	if !strings.Contains(provider.lastReq.SystemInstruction, "nama persona: friendly") {
		t.Fatalf("expected persona marker in system instruction, got: %s", provider.lastReq.SystemInstruction)
	}
	if !strings.Contains(provider.lastReq.SystemInstruction, "signature_id: recova.persona.friendly.v1") {
		t.Fatalf("expected friendly signature in system instruction, got: %s", provider.lastReq.SystemInstruction)
	}
}

func TestPersonaStyleInstruction_HasDistinctSignature(t *testing.T) {
	supportive := personaStyleInstruction("supportive")
	friendly := personaStyleInstruction("friendly")
	direct := personaStyleInstruction("direct")
	concise := personaStyleInstruction("concise")

	if !strings.Contains(supportive, "signature_id: recova.persona.supportive.v1") || !strings.Contains(supportive, "validasi emosi") {
		t.Fatalf("supportive signature missing: %s", supportive)
	}
	if !strings.Contains(friendly, "signature_id: recova.persona.friendly.v1") || !strings.Contains(friendly, "sapaan santai") {
		t.Fatalf("friendly signature missing: %s", friendly)
	}
	if !strings.Contains(direct, "signature_id: recova.persona.direct.v1") || !strings.Contains(direct, "langkah bernomor") {
		t.Fatalf("direct signature missing: %s", direct)
	}
	if !strings.Contains(concise, "signature_id: recova.persona.concise.v1") || !strings.Contains(concise, "maksimal 3 kalimat") {
		t.Fatalf("concise signature missing: %s", concise)
	}
	if supportive == friendly || friendly == direct || direct == concise {
		t.Fatal("persona signature must be distinct across personas")
	}
}

func TestMapAIHistory_BudgetEnforced_TruncatesOlderMessages(t *testing.T) {
	// Create messages: m1 (user, 3000), m2 (assistant, 3000), m3 (user, 2000), m4 (assistant, 2000)
	// Total: 10000 chars > maxCoachHistoryChars (8000)
	// Should keep newest user-initiated pair: m3 (user) + m4 (assistant) = 4000 chars
	m1 := models.AIChat{Role: "user", Content: strings.Repeat("a", 3000)}
	m2 := models.AIChat{Role: "assistant", Content: strings.Repeat("b", 3000)}
	m3 := models.AIChat{Role: "user", Content: strings.Repeat("c", 2000)}
	m4 := models.AIChat{Role: "assistant", Content: strings.Repeat("d", 2000)}

	res := mapAIHistory([]models.AIChat{m1, m2, m3, m4})
	if len(res) != 2 {
		t.Fatalf("expected 2 messages kept within budget, got %d", len(res))
	}
	if res[0].Content != m3.Content || res[0].Role != "user" {
		t.Fatalf("expected first kept message to be m3 user turn")
	}
	if res[1].Content != m4.Content || res[1].Role != "assistant" {
		t.Fatalf("expected second kept message to be m4 assistant turn")
	}
}

func TestMapAIHistory_EnsuresHistoryStartsWithUserTurn(t *testing.T) {
	// If truncation would land on assistant turn:
	// m1 (user, 5000), m2 (assistant, 4000), m3 (user, 2000), m4 (assistant, 1000)
	// Budget (8000) fits m2+m3+m4 (7000 chars), but m2 is assistant.
	// Must drop m2 so history starts with user turn (m3 + m4 = 3000 chars).
	m1 := models.AIChat{Role: "user", Content: strings.Repeat("a", 5000)}
	m2 := models.AIChat{Role: "assistant", Content: strings.Repeat("b", 4000)}
	m3 := models.AIChat{Role: "user", Content: strings.Repeat("c", 2000)}
	m4 := models.AIChat{Role: "assistant", Content: strings.Repeat("d", 1000)}

	res := mapAIHistory([]models.AIChat{m1, m2, m3, m4})
	if len(res) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(res))
	}
	if res[0].Role != "user" {
		t.Fatalf("expected first message to have user role, got: %s", res[0].Role)
	}
}

func TestMapAIHistory_SingleHugeMessage_TruncatedOrDropped(t *testing.T) {
	huge := models.AIChat{Role: "user", Content: strings.Repeat("x", 10000)}
	res := mapAIHistory([]models.AIChat{huge})
	// Single message of 10k chars exceeds 8k budget -> keepFrom moves past it -> returns nil
	if len(res) != 0 {
		t.Fatalf("expected 0 messages when single turn exceeds entire budget, got %d", len(res))
	}
}

func ptrString(v string) *string {
	return &v
}

type fakeAIRepo struct {
	user            models.User
	userErr         error
	profile         models.Profile
	profileErr      error
	activeStreak    *models.Streak
	activeStreakErr error
	history         []models.AIChat
	historyErr      error
	createdMessages []models.AIChat
	createErr       error
	persona         models.UserAIPersonaPreference
	personaErr      error
	upsertErr       error
	upsertPersona   models.UserAIPersonaPreference
}

func (r *fakeAIRepo) GetCoachContext(_ context.Context, _ string, _ int) (CoachContext, error) {
	if r.userErr != nil {
		return CoachContext{}, r.userErr
	}
	if strings.TrimSpace(r.user.ID) == "" {
		return CoachContext{}, gorm.ErrRecordNotFound
	}
	if r.profileErr != nil && !IsRecordNotFound(r.profileErr) {
		return CoachContext{}, r.profileErr
	}
	if r.activeStreakErr != nil {
		return CoachContext{}, r.activeStreakErr
	}
	if r.historyErr != nil {
		return CoachContext{}, r.historyErr
	}
	if r.personaErr != nil && !IsRecordNotFound(r.personaErr) {
		return CoachContext{}, r.personaErr
	}

	persona := DefaultPersona
	if r.persona.Persona != "" {
		persona = ResolvePersonaOrDefault(r.persona.Persona)
	}

	return CoachContext{
		User:         r.user,
		Profile:      r.profile,
		HasProfile:   r.profile.ID != "",
		ActiveStreak: r.activeStreak,
		RecentChats:  r.history,
		Persona:      persona,
	}, nil
}

func (r *fakeAIRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.userErr != nil {
		return models.User{}, r.userErr
	}
	if strings.TrimSpace(r.user.ID) == "" {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *fakeAIRepo) FindProfileByUserID(_ context.Context, _ string) (models.Profile, error) {
	if r.profileErr != nil {
		return models.Profile{}, r.profileErr
	}
	return r.profile, nil
}

func (r *fakeAIRepo) FindActiveStreakByUserID(_ context.Context, _ string) (*models.Streak, error) {
	if r.activeStreakErr != nil {
		return nil, r.activeStreakErr
	}
	return r.activeStreak, nil
}

func (r *fakeAIRepo) ListRecentChatsByUserID(_ context.Context, _ string, _ int) ([]models.AIChat, error) {
	if r.historyErr != nil {
		return nil, r.historyErr
	}
	return r.history, nil
}

func (r *fakeAIRepo) CreateChatMessages(_ context.Context, rows []models.AIChat) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.createdMessages = rows
	return nil
}

func (r *fakeAIRepo) GetPersonaPreferenceByUserID(_ context.Context, _ string) (models.UserAIPersonaPreference, error) {
	if r.personaErr != nil {
		return models.UserAIPersonaPreference{}, r.personaErr
	}
	if strings.TrimSpace(r.persona.UserID) == "" {
		return models.UserAIPersonaPreference{}, gorm.ErrRecordNotFound
	}
	return r.persona, nil
}

func (r *fakeAIRepo) UpsertPersonaPreference(_ context.Context, userID string, persona string, updatedAt time.Time) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upsertPersona = models.UserAIPersonaPreference{
		UserID:    strings.TrimSpace(userID),
		Persona:   strings.TrimSpace(persona),
		UpdatedAt: updatedAt,
	}
	return nil
}

type fakeAIProvider struct {
	response aiplatform.GenerateResponse
	err      error
	lastReq  aiplatform.GenerateRequest
}

func (p *fakeAIProvider) Generate(_ context.Context, req aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	p.lastReq = req
	if p.err != nil {
		return aiplatform.GenerateResponse{}, p.err
	}
	if strings.TrimSpace(p.response.Text) == "" {
		return aiplatform.GenerateResponse{}, errors.New("missing response")
	}
	return p.response, nil
}

type mockTelemetry struct {
	recordedActions  []string
	recordedPersonas []string
	recordedErrors   []error
}

func (m *mockTelemetry) RecordPersonaUsage(action string, persona string, err error) {
	m.recordedActions = append(m.recordedActions, action)
	m.recordedPersonas = append(m.recordedPersonas, persona)
	m.recordedErrors = append(m.recordedErrors, err)
}

func TestSanitizeCoachReply_MarkerVariants(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "Halo, semangat terus ya!\n\nsignature_id: recova.persona.supportive.v1",
			expected: "Halo, semangat terus ya!",
		},
		{
			input:    "Halo!\n- signature_id: recova.persona.friendly.v1\nAda yang bisa dibantu?",
			expected: "Halo!\nAda yang bisa dibantu?",
		},
		{
			input:    "Langkah awal sudah bagus recova.persona.direct.v1.",
			expected: "Langkah awal sudah bagus",
		},
		{
			input:    "Tetap fokus hari ini recova.persona.concise.v1,",
			expected: "Tetap fokus hari ini",
		},
		{
			input:    "signature_id: recova.persona.supportive.v1",
			expected: "",
		},
	}

	for _, tc := range cases {
		got := sanitizeCoachReply(tc.input)
		if got != tc.expected {
			t.Errorf("sanitizeCoachReply(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestService_AskCoach_RepositoryErrors(t *testing.T) {
	// 1. User not found
	svc1 := NewService(&fakeAIRepo{userErr: gorm.ErrRecordNotFound}, &fakeAIProvider{})
	_, err := svc1.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected not found for missing user, got %v", err)
	}

	// 2. User repo error
	svc2 := NewService(&fakeAIRepo{userErr: errors.New("db error")}, &fakeAIProvider{})
	_, err = svc2.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error for user repo error, got %v", err)
	}

	// 3. Profile repo error
	repo3 := &fakeAIRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		profileErr: errors.New("db error"),
	}
	svc3 := NewService(repo3, &fakeAIProvider{})
	_, err = svc3.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error for profile repo error, got %v", err)
	}

	// 4. Streak repo error
	repo4 := &fakeAIRepo{
		user:            models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		activeStreakErr: errors.New("db error"),
	}
	svc4 := NewService(repo4, &fakeAIProvider{})
	_, err = svc4.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error for streak repo error, got %v", err)
	}

	// 5. History repo error
	repo5 := &fakeAIRepo{
		user:       models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		historyErr: errors.New("db error"),
	}
	svc5 := NewService(repo5, &fakeAIProvider{})
	_, err = svc5.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error for history repo error, got %v", err)
	}
}

func TestService_AskCoach_ChatPersistenceError(t *testing.T) {
	repo := &fakeAIRepo{
		user:      models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		createErr: errors.New("insert failed"),
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Saran AI"}}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error on chat persistence error, got %v", err)
	}
}

func TestService_AskCoach_Cancellation(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	provider := &fakeAIProvider{err: context.Canceled}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
}

func TestService_AskCoach_EmptyProviderResponse(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
	}
	provider := &fakeAIProvider{
		err: &aiplatform.ProviderError{
			Kind:    aiplatform.ErrorKindInvalidResponse,
			Message: "provider response content is empty",
		},
	}
	service := NewService(repo, provider)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err == nil || errs.Map(err).Code != errs.CodeDownstreamError {
		t.Fatalf("expected downstream error on empty response, got %v", err)
	}
}

func TestService_GetSummary_RepositoryErrors(t *testing.T) {
	// 1. User repo error
	svc1 := NewService(&fakeAIRepo{userErr: errors.New("db error")}, &fakeAIProvider{})
	_, err := svc1.GetSummary(context.Background(), "user-1")
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error for user repo error, got %v", err)
	}

	// 2. Profile repo error
	svc2 := NewService(&fakeAIRepo{
		user:       models.User{ID: "user-1", Email: "u@test.local"},
		profileErr: errors.New("db error"),
	}, &fakeAIProvider{})
	_, err = svc2.GetSummary(context.Background(), "user-1")
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error for profile repo error, got %v", err)
	}
}

func TestService_GenerateRelapseSolution_InvalidJSON(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Email: "user@test.local"},
	}
	provider := &fakeAIProvider{
		response: aiplatform.GenerateResponse{Text: "this is clearly not json"},
	}
	service := NewService(repo, provider)

	_, err := service.GenerateRelapseSolution(context.Background(), "user-1", RelapseSolutionRequest{
		Mood:           "cemas",
		RelapseTrigger: []string{"stres"},
	})
	if err == nil {
		t.Fatal("expected error on invalid json, got nil")
	}
	if errs.Map(err).Code != errs.CodeDownstreamError {
		t.Fatalf("expected downstream error on invalid json, got %v", errs.Map(err).Code)
	}
}

func TestService_GenerateRelapseSolution_ProviderError(t *testing.T) {
	repo := &fakeAIRepo{
		user: models.User{ID: "user-1", Email: "user@test.local"},
	}
	provider := &fakeAIProvider{
		err: errors.New("provider unavailable"),
	}
	service := NewService(repo, provider)

	_, err := service.GenerateRelapseSolution(context.Background(), "user-1", RelapseSolutionRequest{
		Mood:           "cemas",
		RelapseTrigger: []string{"stres"},
	})
	if err == nil {
		t.Fatal("expected error on provider failure, got nil")
	}
	if errs.Map(err).Code != errs.CodeDownstreamError {
		t.Fatalf("expected downstream error on provider error, got %v", errs.Map(err).Code)
	}
}

func TestService_GetPersonaPreference_RepositoryError(t *testing.T) {
	repo := &fakeAIRepo{
		user:       models.User{ID: "user-1"},
		personaErr: errors.New("db error"),
	}
	service := NewService(repo, &fakeAIProvider{})

	_, err := service.GetPersonaPreference(context.Background(), "user-1")
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error on persona repo error, got %v", err)
	}
}

func TestService_UpdatePersonaPreference_RepositoryError(t *testing.T) {
	repo := &fakeAIRepo{
		user:      models.User{ID: "user-1"},
		upsertErr: errors.New("db error"),
	}
	service := NewService(repo, &fakeAIProvider{})

	_, err := service.UpdatePersonaPreference(context.Background(), "user-1", PersonaPreferenceRequest{
		Persona: "supportive",
	})
	if err == nil || errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected internal error on persona upsert error, got %v", err)
	}
}

func TestService_Telemetry_NilDoesNotPanic(t *testing.T) {
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		persona: models.UserAIPersonaPreference{UserID: "user-1", Persona: "supportive"},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Jawaban"}}
	service := NewService(repo, provider)

	// Explicitly pass nil to SetTelemetry
	service.SetTelemetry(nil)

	// Ensure calls don't panic
	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = service.GetPersonaPreference(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = service.UpdatePersonaPreference(context.Background(), "user-1", PersonaPreferenceRequest{
		Persona: "friendly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Telemetry_RecordsPersonaUsage(t *testing.T) {
	repo := &fakeAIRepo{
		user:    models.User{ID: "user-1", Nickname: "tester", Email: "user@example.test"},
		persona: models.UserAIPersonaPreference{UserID: "user-1", Persona: "supportive"},
	}
	provider := &fakeAIProvider{response: aiplatform.GenerateResponse{Text: "Jawaban"}}
	service := NewService(repo, provider)

	tel := &mockTelemetry{}
	service.SetTelemetry(tel)

	_, err := service.AskCoach(context.Background(), "user-1", AskCoachRequest{Message: "halo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = service.GetPersonaPreference(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = service.UpdatePersonaPreference(context.Background(), "user-1", PersonaPreferenceRequest{
		Persona: "direct",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedActions := []string{"ask_coach", "persona_preference_get", "persona_preference_update"}
	if len(tel.recordedActions) != len(expectedActions) {
		t.Fatalf("expected %d telemetry records, got %d", len(expectedActions), len(tel.recordedActions))
	}
	for i, expected := range expectedActions {
		if tel.recordedActions[i] != expected {
			t.Errorf("action[%d] = %s, expected %s", i, tel.recordedActions[i], expected)
		}
	}
}
