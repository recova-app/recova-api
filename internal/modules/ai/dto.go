package ai

// AskCoachRequest is request payload for AI coach chat.
type AskCoachRequest struct {
	Message string `json:"message"`
}

// AskCoachResponseData is success data for ask-coach endpoint.
type AskCoachResponseData struct {
	Response    string `json:"response"`
	PersonaUsed string `json:"persona_used"`
}

// ChatHistoryQuery captures optional chat-history query parameters.
type ChatHistoryQuery struct {
	Limit *int `query:"limit"`
}

// ChatHistoryItem is one chat-history record payload.
type ChatHistoryItem struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// SummaryResponseData is success data for summary endpoint.
type SummaryResponseData struct {
	Summary string `json:"summary"`
}

// OnboardingAnalysisRequest is request payload for onboarding analysis.
type OnboardingAnalysisRequest struct {
	Answers map[string]any `json:"answers"`
}

// PersonaPreferenceRequest is request payload for persona preference update.
type PersonaPreferenceRequest struct {
	Persona string `json:"persona"`
}

// RelapseSolutionRequest is request payload for relapse solution generation.
type RelapseSolutionRequest struct {
	Mood           string   `json:"mood"`
	RelapseTrigger []string `json:"relapse_trigger"`
	Commitment     *string  `json:"commitment"`
}

// PersonaPreferenceResponseData is success data for persona preference endpoints.
type PersonaPreferenceResponseData struct {
	Persona         string `json:"persona"`
	FallbackPersona string `json:"fallback_persona"`
}

// OnboardingAnalysisResponseData is success data for onboarding-analysis endpoint.
type OnboardingAnalysisResponseData struct {
	Level            string `json:"level"`
	Title            string `json:"title"`
	LevelDescription string `json:"level_description"`
	PatternAnalysis  string `json:"pattern_analysis"`
	Encouragement    string `json:"encouragement"`
}

// RelapseSolutionResponseData is success data for relapse-solution endpoint.
type RelapseSolutionResponseData struct {
	Title       string   `json:"title"`
	Analysis    string   `json:"analysis"`
	ActionSteps []string `json:"action_steps"`
}

// AskCoachInput is normalized ask-coach payload.
type AskCoachInput struct {
	Message string
}

// OnboardingAnalysisInput is normalized onboarding-analysis payload.
type OnboardingAnalysisInput struct {
	Answers map[string]any
}

// PersonaPreferenceInput is normalized persona preference update payload.
type PersonaPreferenceInput struct {
	Persona string
}

// RelapseSolutionInput is normalized relapse-solution payload.
type RelapseSolutionInput struct {
	Mood           string
	RelapseTrigger []string
	Commitment     *string
}
