package education

import "time"

// ContentPayload is one education content item for API response.
type ContentPayload struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	URL          string  `json:"url"`
	ThumbnailURL *string `json:"thumbnail_url"`
	Category     string  `json:"category"`
	Type         string  `json:"type"`
	PublishedAt  *string `json:"published_at"`
}

func formatPublishedAt(ts *time.Time) *string {
	if ts == nil {
		return nil
	}
	formatted := ts.UTC().Format(time.RFC3339)
	return &formatted
}
