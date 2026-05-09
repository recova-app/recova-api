package ai

// AskCoachRequest is request payload for AI coach chat.
type AskCoachRequest struct {
	Message string `json:"message"`
}

// AskCoachResponseData is success data for ask-coach endpoint.
type AskCoachResponseData struct {
	Response    string `json:"response"`
	PersonaUsed string `json:"personaUsed"`
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
	CreatedAt string `json:"createdAt"`
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

// PersonaPreferenceResponseData is success data for persona preference endpoints.
type PersonaPreferenceResponseData struct {
	Persona         string `json:"persona"`
	FallbackPersona string `json:"fallbackPersona"`
}

// OnboardingAnalysisResponseData is success data for onboarding-analysis endpoint.
type OnboardingAnalysisResponseData struct {
	Level            string `json:"level"`
	Title            string `json:"title"`
	LevelDescription string `json:"levelDescription"`
	PatternAnalysis  string `json:"patternAnalysis"`
	Encouragement    string `json:"encouragement"`
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
