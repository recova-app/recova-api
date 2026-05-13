package models

import "time"

// AuthRefreshToken stores hashed refresh-token state for rotation and revocation.
type AuthRefreshToken struct {
	ID            string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        string     `gorm:"type:uuid;column:user_id;not null;index:idx_auth_refresh_tokens_user_id"`
	TokenHash     string     `gorm:"column:token_hash;not null;uniqueIndex"`
	ExpiresAt     time.Time  `gorm:"column:expires_at;not null"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	RotatedFromID *string    `gorm:"type:uuid;column:rotated_from_id"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:now()"`
}

// TableName defines explicit table name for GORM mapping.
func (AuthRefreshToken) TableName() string {
	return "auth_refresh_tokens"
}
