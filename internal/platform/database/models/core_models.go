package models

import (
	"time"
)

type User struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GoogleID    string    `gorm:"column:google_id;not null;uniqueIndex"`
	Email       string    `gorm:"not null;uniqueIndex"`
	Nickname    string    `gorm:"not null"`
	UserWhy     *string   `gorm:"column:user_why"`
	CheckInTime *string   `gorm:"column:check_in_time;type:time"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
}

type Profile struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          string    `gorm:"type:uuid;column:user_id;not null;uniqueIndex"`
	Answers         []byte    `gorm:"type:jsonb;not null"`
	DependencyLevel *string   `gorm:"column:dependency_level"`
	AISummary       *string   `gorm:"column:ai_summary"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`
	UpdatedAt       time.Time `gorm:"not null;default:now()"`
}

type CheckIn struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       string    `gorm:"type:uuid;column:user_id;not null;index"`
	CheckInDate  time.Time `gorm:"column:check_in_date;type:date;not null"`
	Mood         string    `gorm:"not null"`
	IsSuccessful bool      `gorm:"column:is_successful;not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

type Streak struct {
	ID        string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string     `gorm:"type:uuid;column:user_id;not null;index"`
	StartDate time.Time  `gorm:"column:start_date;not null"`
	EndDate   *time.Time `gorm:"column:end_date"`
	IsActive  bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
	UpdatedAt time.Time  `gorm:"not null;default:now()"`
}

type Journal struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;column:user_id;not null;index"`
	CheckInID *string   `gorm:"type:uuid;column:check_in_id;uniqueIndex"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

type CommunityPost struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       string `gorm:"type:uuid;column:user_id;not null;index"`
	Title        *string
	Content      string    `gorm:"not null"`
	Category     string    `gorm:"not null;default:advice"`
	CommentCount int       `gorm:"column:comment_count;not null;default:0"`
	LikeCount    int       `gorm:"column:like_count;not null;default:0"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}

type CommunityComment struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;column:user_id;not null;index:idx_community_comments_user_post"`
	PostID    string    `gorm:"type:uuid;column:post_id;not null;index:idx_community_comments_user_post"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

type CommunityPostLike struct {
	UserID    string    `gorm:"type:uuid;column:user_id;not null;primaryKey"`
	PostID    string    `gorm:"type:uuid;column:post_id;not null;primaryKey;index"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

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

type DailyMotivation struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Content   string    `gorm:"not null;uniqueIndex"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

type DailyChallenge struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Content   string    `gorm:"not null;uniqueIndex"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

type AIChat struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `gorm:"type:uuid;column:user_id;not null;index"`
	Role      string    `gorm:"not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now();index:idx_ai_chats_user_created_at"`
}
