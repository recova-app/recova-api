package achievements

// CategoryQuery captures optional achievement category filter.
type CategoryQuery struct {
	Category *string `query:"category"`
}

// CatalogResponse is API payload for achievement catalog endpoint.
type CatalogResponse struct {
	Items []CatalogItem `json:"items"`
}

// CatalogItem describes one achievement from catalog.
type CatalogItem struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Threshold   float64 `json:"threshold"`
}

// ProgressResponse is API payload for achievement progress endpoint.
type ProgressResponse struct {
	Items []ProgressItem `json:"items"`
}

// ProgressItem describes one achievement progress row.
type ProgressItem struct {
	AchievementCode string  `json:"achievementCode"`
	Category        string  `json:"category"`
	Threshold       float64 `json:"threshold"`
	ProgressValue   float64 `json:"progressValue"`
	Unlocked        bool    `json:"unlocked"`
	UnlockedAt      *string `json:"unlockedAt,omitempty"`
}

// UnlockedResponse is API payload for unlocked achievements endpoint.
type UnlockedResponse struct {
	Items []UnlockedItem `json:"items"`
}

// UnlockedItem describes one unlocked achievement.
type UnlockedItem struct {
	AchievementCode string  `json:"achievementCode"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Category        string  `json:"category"`
	Threshold       float64 `json:"threshold"`
	ProgressValue   float64 `json:"progressValue"`
	UnlockedAt      string  `json:"unlockedAt"`
}
