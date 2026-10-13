package models

// RegisteredModels returns pointer instances of all domain models for migration drift checking.
func RegisteredModels() []any {
	return []any{
		&User{},
		&Profile{},
		&Streak{},
		&CheckIn{},
		&Journal{},
		&Relapse{},
		&CommunityPost{},
		&CommunityComment{},
		&CommunityPostLike{},
		&Achievement{},
		&UserAchievementProgress{},
		&EducationContent{},
		&DailyMotivation{},
		&DailyChallenge{},
		&DailyPhysicalChallenge{},
		&AIChat{},
		&UserAIPersonaPreference{},
		&AuthRefreshToken{},
	}
}
