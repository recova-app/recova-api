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
	CurrentStreak           int                `json:"currentStreak"`
	LongestStreak           int                `json:"longestStreak"`
	TotalCheckins           int                `json:"totalCheckins"`
	StreakCalendar          []string           `json:"streakCalendar"`
	RelapseCount            int                `json:"relapseCount"`
	RelapseRate             float64            `json:"relapseRate"`
	RecoverySuccessRate     float64            `json:"recoverySuccessRate"`
	CheckinConsistencyScore float64            `json:"checkinConsistencyScore"`
	WeeklyProgress          ProgressPayload    `json:"weeklyProgress"`
	MonthlyProgress         ProgressPayload    `json:"monthlyProgress"`
	MoodTrend               []MoodTrendPayload `json:"moodTrend"`
}

// ProgressPayload is periodic progress summary payload.
type ProgressPayload struct {
	WindowDays                 int     `json:"windowDays"`
	CurrentSuccessfulCheckins  int     `json:"currentSuccessfulCheckins"`
	PreviousSuccessfulCheckins int     `json:"previousSuccessfulCheckins"`
	Delta                      int     `json:"delta"`
	DeltaRate                  float64 `json:"deltaRate"`
}

// MoodTrendPayload is mood trend bucket payload.
type MoodTrendPayload struct {
	Date            string  `json:"date"`
	DominantMood    string  `json:"dominantMood"`
	SuccessfulRatio float64 `json:"successfulRatio"`
}

// ActivitySummaryQuery captures query parameters for periodic activity summary endpoint.
type ActivitySummaryQuery struct {
	WindowDays *int `query:"windowDays"`
}

// ActivitySummaryPayload is summary payload for periodic activity endpoint.
type ActivitySummaryPayload struct {
	WindowDays         int                   `json:"windowDays"`
	SuccessfulCheckins int                   `json:"successfulCheckins"`
	Relapses           int                   `json:"relapses"`
	ActiveDays         int                   `json:"activeDays"`
	RecentActivity     []ActivityItemPayload `json:"recentActivity"`
}

// ActivityItemPayload is one activity timeline item.
type ActivityItemPayload struct {
	Date string  `json:"date"`
	Type string  `json:"type"`
	Mood *string `json:"mood,omitempty"`
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
