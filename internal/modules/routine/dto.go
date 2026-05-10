package routine

// DailyCheckInRequest is request payload for daily check-in submission.
type DailyCheckInRequest struct {
	Mood         string  `json:"mood"`
	IsSuccessful *bool   `json:"is_successful"`
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
	UserID       string  `json:"user_id"`
	CheckInDate  string  `json:"check_in_date"`
	Mood         string  `json:"mood"`
	IsSuccessful bool    `json:"is_successful"`
	Commitment   *string `json:"commitment"`
	CreatedAt    string  `json:"created_at"`
}

// StatisticsPayload is API payload for routine statistics.
type StatisticsPayload struct {
	CurrentStreak           int                `json:"current_streak"`
	LongestStreak           int                `json:"longest_streak"`
	TotalCheckins           int                `json:"total_checkins"`
	StreakCalendar          []string           `json:"streak_calendar"`
	RelapseCount            int                `json:"relapse_count"`
	RelapseRate             float64            `json:"relapse_rate"`
	RecoverySuccessRate     float64            `json:"recovery_success_rate"`
	CheckinConsistencyScore float64            `json:"checkin_consistency_score"`
	WeeklyProgress          ProgressPayload    `json:"weekly_progress"`
	MonthlyProgress         ProgressPayload    `json:"monthly_progress"`
	MoodTrend               []MoodTrendPayload `json:"mood_trend"`
}

// ProgressPayload is periodic progress summary payload.
type ProgressPayload struct {
	WindowDays                 int     `json:"window_days"`
	CurrentSuccessfulCheckins  int     `json:"current_successful_checkins"`
	PreviousSuccessfulCheckins int     `json:"previous_successful_checkins"`
	Delta                      int     `json:"delta"`
	DeltaRate                  float64 `json:"delta_rate"`
}

// MoodTrendPayload is mood trend bucket payload.
type MoodTrendPayload struct {
	Date            string  `json:"date"`
	DominantMood    string  `json:"dominant_mood"`
	SuccessfulRatio float64 `json:"successful_ratio"`
}

// ActivitySummaryQuery captures query parameters for periodic activity summary endpoint.
type ActivitySummaryQuery struct {
	WindowDays *int `query:"window_days"`
}

// ActivitySummaryPayload is summary payload for periodic activity endpoint.
type ActivitySummaryPayload struct {
	WindowDays         int                   `json:"window_days"`
	SuccessfulCheckins int                   `json:"successful_checkins"`
	Relapses           int                   `json:"relapses"`
	ActiveDays         int                   `json:"active_days"`
	RecentActivity     []ActivityItemPayload `json:"recent_activity"`
}

// ActivityItemPayload is one activity timeline item.
type ActivityItemPayload struct {
	Date string  `json:"date"`
	Type string  `json:"type"`
	Mood *string `json:"mood,omitempty"`
}

// CheckInResponseData combines check-in detail and current statistics.
type CheckInResponseData struct {
	CheckIn    CheckInPayload    `json:"check_in"`
	Statistics StatisticsPayload `json:"statistics"`
}

// RelapsePayload is API payload for one relapse history record.
type RelapsePayload struct {
	CheckInID   string  `json:"check_in_id"`
	CheckInDate string  `json:"check_in_date"`
	Mood        string  `json:"mood"`
	Commitment  *string `json:"commitment"`
	CreatedAt   string  `json:"created_at"`
}
