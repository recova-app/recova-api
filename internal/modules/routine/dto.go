package routine

// DailyCheckInRequest is request payload for daily check-in submission.
type DailyCheckInRequest struct {
	Mood         string  `json:"mood"`
	IsSuccessful *bool   `json:"isSuccessful"`
	Commitment   *string `json:"commitment"`
	Content      *string `json:"content"`
}

// DailyCheckInInput is normalized check-in input used by service layer.
type DailyCheckInInput struct {
	Mood         string
	IsSuccessful bool
	JournalText  *string
}

// CheckInPayload is API payload for stored check-in record.
type CheckInPayload struct {
	ID           string  `json:"id"`
	UserID       string  `json:"userId"`
	CheckInDate  string  `json:"checkInDate"`
	Mood         string  `json:"mood"`
	IsSuccessful bool    `json:"isSuccessful"`
	Commitment   *string `json:"commitment"`
	CreatedAt    string  `json:"createdAt"`
}

// StatisticsPayload is API payload for routine statistics.
type StatisticsPayload struct {
	CurrentStreak  int      `json:"currentStreak"`
	LongestStreak  int      `json:"longestStreak"`
	TotalCheckins  int      `json:"totalCheckins"`
	StreakCalendar []string `json:"streakCalendar"`
}

// CheckInResponseData combines check-in detail and current statistics.
type CheckInResponseData struct {
	CheckIn    CheckInPayload    `json:"checkIn"`
	Statistics StatisticsPayload `json:"statistics"`
}

// RelapsePayload is API payload for one relapse history record.
type RelapsePayload struct {
	CheckInID   string  `json:"checkInId"`
	CheckInDate string  `json:"checkInDate"`
	Mood        string  `json:"mood"`
	Commitment  *string `json:"commitment"`
	CreatedAt   string  `json:"createdAt"`
}
