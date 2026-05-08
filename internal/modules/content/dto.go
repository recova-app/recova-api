package content

// DailyContentPayload is API payload for one daily content response.
type DailyContentPayload struct {
	Date       string `json:"date"`
	Motivation string `json:"motivation"`
	Challenge  string `json:"challenge"`
}
