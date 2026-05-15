package content

// DailyContentPayload is API payload for one daily content response.
type DailyContentPayload struct {
	Date              string                `json:"date"`
	Motivation        string                `json:"motivation"`
	Challenge         DailyChallengePayload `json:"challenge"`
	PhysicalChallenge DailyChallengePayload `json:"physical_challenge"`
}

// DailyChallengePayload is payload shape for one challenge item.
type DailyChallengePayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
