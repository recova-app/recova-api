package models

import (
	"time"
)

// User stores core account identity and onboarding pointers.
type User struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GoogleID     *string   `gorm:"column:google_id;uniqueIndex"`
	Email        string    `gorm:"not null;uniqueIndex:uq_users_email"`
	Username     *string   `gorm:"column:username;uniqueIndex:uq_users_username"`
	PasswordHash *string   `gorm:"column:password_hash"`
	Nickname     string    `gorm:"not null"`
	UserWhy      *string   `gorm:"column:user_why"`
	CheckInTime  *string   `gorm:"column:check_in_time;type:time"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}

// Profile stores onboarding profile answers and derived AI summary for one user.
type Profile struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          string    `gorm:"type:uuid;column:user_id;not null;uniqueIndex"`
	Answers         []byte    `gorm:"type:jsonb;not null"`
	DependencyLevel *string   `gorm:"column:dependency_level"`
	AISummary       *string   `gorm:"column:ai_summary"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`
	UpdatedAt       time.Time `gorm:"not null;default:now()"`
}

// CheckIn stores one daily check-in row per user and date.
type CheckIn struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       string    `gorm:"type:uuid;column:user_id;not null;index"`
	CheckInDate  time.Time `gorm:"column:check_in_date;type:date;not null"`
	Mood         string    `gorm:"not null"`
	IsSuccessful bool      `gorm:"column:is_successful;not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

// Streak stores active and historical streak windows for one user.
type Streak struct {
	ID        string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string     `gorm:"type:uuid;column:user_id;not null;index"`
	StartDate time.Time  `gorm:"column:start_date;not null"`
	EndDate   *time.Time `gorm:"column:end_date"`
	IsActive  bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
	UpdatedAt time.Time  `gorm:"not null;default:now()"`
}

// Journal stores one personal journal entry optionally linked to a check-in.
type Journal struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;column:user_id;not null;index"`
	CheckInID *string   `gorm:"type:uuid;column:check_in_id;uniqueIndex"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// CommunityPost stores a public community post created by one user.
type CommunityPost struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       string `gorm:"type:uuid;column:user_id;not null;index"`
	Title        *string
	Content      string    `gorm:"not null"`
	Category     string    `gorm:"not null;default:saran"`
	CommentCount int       `gorm:"column:comment_count;not null;default:0"`
	LikeCount    int       `gorm:"column:like_count;not null;default:0"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}

// CommunityComment stores threaded comments on community posts.
type CommunityComment struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          string    `gorm:"type:uuid;column:user_id;not null;index:idx_community_comments_user_post"`
	PostID          string    `gorm:"type:uuid;column:post_id;not null;index:idx_community_comments_user_post;index:idx_community_comments_post_parent_created,priority:1"`
	ParentCommentID *string   `gorm:"type:uuid;column:parent_comment_id;index:idx_community_comments_parent;index:idx_community_comments_post_parent_created,priority:2"`
	Content         string    `gorm:"not null"`
	Depth           int16     `gorm:"type:smallint;not null;default:0"`
	ReplyCount      int       `gorm:"column:reply_count;not null;default:0"`
	CreatedAt       time.Time `gorm:"not null;default:now();index:idx_community_comments_post_parent_created,priority:3"`
}

// CommunityPostLike stores one like relation between user and post.
type CommunityPostLike struct {
	UserID    string    `gorm:"type:uuid;column:user_id;not null;primaryKey"`
	PostID    string    `gorm:"type:uuid;column:post_id;not null;primaryKey;index"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// Achievement stores achievement catalog metadata and activation state.
type Achievement struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string    `gorm:"not null;uniqueIndex"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"not null"`
	Category    string    `gorm:"not null"`
	Threshold   float64   `gorm:"type:numeric;not null"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

// UserAchievementProgress stores per-user progress state for one achievement.
type UserAchievementProgress struct {
	ID              string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          string     `gorm:"type:uuid;column:user_id;not null;index:idx_user_achievement_progress_user,priority:1;index:idx_user_achievement_progress_user_unlocked,priority:1;uniqueIndex:uq_user_achievement_progress_user_achievement,priority:1"`
	AchievementID   string     `gorm:"type:uuid;column:achievement_id;not null;index:idx_user_achievement_progress_achievement;uniqueIndex:uq_user_achievement_progress_user_achievement,priority:2"`
	ProgressValue   float64    `gorm:"column:progress_value;type:numeric;not null;default:0"`
	UnlockedAt      *time.Time `gorm:"column:unlocked_at;index:idx_user_achievement_progress_user_unlocked,priority:2"`
	LastEvaluatedAt time.Time  `gorm:"column:last_evaluated_at;not null;default:now()"`
	CreatedAt       time.Time  `gorm:"not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"not null;default:now();index:idx_user_achievement_progress_user,priority:2,sort:desc"`
}

// TableName maps achievement progress model to legacy singular table name.
func (UserAchievementProgress) TableName() string {
	return "user_achievement_progress"
}

// EducationContent stores one educational content item metadata.
type EducationContent struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string `gorm:"not null"`
	Description  *string
	URL          string     `gorm:"not null"`
	ThumbnailURL *string    `gorm:"column:thumbnail_url"`
	Category     string     `gorm:"not null"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	PublishedAt  *time.Time `gorm:"column:published_at"`
}

// DailyMotivation stores one daily motivation catalog item.
type DailyMotivation struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Content   string    `gorm:"not null;uniqueIndex"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// DailyChallenge stores one daily challenge catalog item.
type DailyChallenge struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Content   string    `gorm:"not null;uniqueIndex"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// AIChat stores one AI conversation message for one user.
type AIChat struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;column:user_id;not null;index"`
	Role      string    `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now();index:idx_ai_chats_user_created_at"`
}

// UserAIPersonaPreference stores selected AI persona preference for one user.
type UserAIPersonaPreference struct {
	UserID    string    `gorm:"type:uuid;column:user_id;primaryKey"`
	Persona   string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now()"`
}

func (UserAIPersonaPreference) TableName() string {
	return "user_ai_persona_preferences"
}
