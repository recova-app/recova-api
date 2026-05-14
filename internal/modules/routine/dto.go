package routine

// DailyCheckInRequest is request payload for daily check-in submission.
type DailyCheckInRequest struct {
	Mood           string   `json:"mood"`
	IsSuccessful   *bool    `json:"is_successful"`
	Commitment     *string  `json:"commitment"`
	Content        *string  `json:"content"`
	RelapseTrigger []string `json:"relapse_trigger"`
}

// DailyCheckInInput is normalized check-in input used by service layer.
type DailyCheckInInput struct {
	Mood         string
	IsSuccessful bool
	JournalText  *string
}

// RelapseRequest is request payload for explicit relapse submission.
type RelapseRequest struct {
	Mood           string   `json:"mood"`
	RelapseTrigger []string `json:"relapse_trigger"`
	Commitment     *string  `json:"commitment"`
	Content        *string  `json:"content"`
}

// RelapseInput is normalized relapse input used by service layer.
type RelapseInput struct {
	Mood           string
	RelapseTrigger []string
	JournalText    *string
}

// CheckInPayload is API payload for stored check-in record.
type CheckInPayload struct {
	ID             string   `json:"id"`
	UserID         string   `json:"user_id"`
	CheckInDate    string   `json:"check_in_date"`
	CheckInDayName string   `json:"check_in_day_name"`
	Mood           string   `json:"mood"`
	IsSuccessful   bool     `json:"is_successful"`
	Commitment     *string  `json:"commitment"`
	RelapseTrigger []string `json:"relapse_trigger"`
	CreatedAt      string   `json:"created_at"`
}

// StatisticsPayload is API payload for routine statistics.
type StatisticsPayload struct {
	CurrentStreak           int                         `json:"current_streak"`
	LongestStreak           int                         `json:"longest_streak"`
	TotalCheckins           int                         `json:"total_checkins"`
	TotalAttempts           int                         `json:"total_attempts"`
	SuccessRate             float64                     `json:"success_rate"`
	StreakCalendar          []string                    `json:"streak_calendar"`
	RelapseCount            int                         `json:"relapse_count"`
	RelapseRate             float64                     `json:"relapse_rate"`
	RecoverySuccessRate     float64                     `json:"recovery_success_rate"`
	CheckinConsistencyScore float64                     `json:"checkin_consistency_score"`
	WeeklyProgress          ProgressPayload             `json:"weekly_progress"`
	MonthlyProgress         ProgressPayload             `json:"monthly_progress"`
	MoodTrend               []MoodTrendPayload          `json:"mood_trend"`
	LastCheckInDate         *string                     `json:"last_check_in_date"`
	LastCheckInDayName      *string                     `json:"last_check_in_day_name"`
	LastRelapseDate         *string                     `json:"last_relapse_date"`
	LastRelapseDayName      *string                     `json:"last_relapse_day_name"`
	WeekdaySummary          []WeekdaySummaryPayload     `json:"weekday_summary"`
	StreakGoalComparison    StreakGoalComparisonPayload `json:"streak_goal_comparison"`
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
	DayName         string  `json:"day_name"`
	DominantMood    string  `json:"dominant_mood"`
	SuccessfulRatio float64 `json:"successful_ratio"`
}

// WeekdaySummaryPayload is aggregate check-in summary grouped by day name (Bahasa Indonesia).
type WeekdaySummaryPayload struct {
	DayName            string  `json:"day_name"`
	SuccessfulCheckins int     `json:"successful_checkins"`
	RelapseCount       int     `json:"relapse_count"`
	TotalCheckins      int     `json:"total_checkins"`
	SuccessRate        float64 `json:"success_rate"`
}

// StreakGoalComparisonPayload compares streak progress against user porn_free_goal.
type StreakGoalComparisonPayload struct {
	PornFreeGoal  *int    `json:"porn_free_goal"`
	CurrentStreak int     `json:"current_streak"`
	LongestStreak int     `json:"longest_streak"`
	GoalReached   bool    `json:"goal_reached"`
	RemainingDays *int    `json:"remaining_days"`
	ProgressRate  float64 `json:"progress_rate"`
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
	Date    string  `json:"date"`
	DayName string  `json:"day_name"`
	Type    string  `json:"type"`
	Mood    *string `json:"mood,omitempty"`
}

// CheckInResponseData combines check-in detail and current statistics.
type CheckInResponseData struct {
	CheckIn         CheckInPayload          `json:"check_in"`
	Statistics      StatisticsPayload       `json:"statistics"`
	RelapseSolution *RelapseSolutionPayload `json:"relapse_solution"`
}

// RelapseSolutionPayload is instant AI action plan generated when relapse happens.
type RelapseSolutionPayload struct {
	Title       string   `json:"title"`
	Analysis    string   `json:"analysis"`
	ActionSteps []string `json:"action_steps"`
	GeneratedAt string   `json:"generated_at"`
}

// RelapsePayload is API payload for one relapse history record.
type RelapsePayload struct {
	ID             string   `json:"id"`
	UserID         string   `json:"user_id"`
	RelapseDate    string   `json:"relapse_date"`
	RelapseDayName string   `json:"relapse_day_name"`
	Mood           string   `json:"mood"`
	Commitment     *string  `json:"commitment"`
	RelapseTrigger []string `json:"relapse_trigger"`
	CheckInID      *string  `json:"check_in_id"`
	CreatedAt      string   `json:"created_at"`
}

// RelapseResponseData combines relapse detail and current statistics.
type RelapseResponseData struct {
	Relapse         RelapsePayload          `json:"relapse"`
	Statistics      StatisticsPayload       `json:"statistics"`
	RelapseSolution *RelapseSolutionPayload `json:"relapse_solution"`
}

// RelapseStatisticsResponseData is complete relapse statistics payload.
type RelapseStatisticsResponseData struct {
	Statistics                StatisticsPayload         `json:"statistics"`
	Relapses                  []RelapsePayload          `json:"relapses"`
	HourlyRelapseDistribution []RelapseHourStatPayload  `json:"hourly_relapse_distribution"`
	PeakRelapseHoursUTC       []int                     `json:"peak_relapse_hours_utc"`
	PeakRelapseCount          int                       `json:"peak_relapse_count"`
	AISummary                 string                    `json:"ai_summary"`
	RelapseTimeSummary        RelapseTimeSummaryPayload `json:"relapse_time_summary"`
	LatestRelapseSolution     *RelapseSolutionPayload   `json:"latest_relapse_solution"`
}

// RelapseHourStatPayload is relapse distribution grouped by UTC hour.
type RelapseHourStatPayload struct {
	HourUTC      int `json:"hour_utc"`
	RelapseCount int `json:"relapse_count"`
}

// RelapseTimeSummaryPayload is AI suggestion summary for peak relapse time.
type RelapseTimeSummaryPayload struct {
	Title               string   `json:"title"`
	Summary             string   `json:"summary"`
	SuggestedActivities []string `json:"suggested_activities"`
	GeneratedAt         string   `json:"generated_at"`
}
