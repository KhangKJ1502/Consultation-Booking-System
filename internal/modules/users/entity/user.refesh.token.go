// package entity

// import (
// 	"time"

// 	"github.com/google/uuid"
// )

// // UserRefreshToken represents tbl_user_refresh_tokens table
// type UserRefreshToken struct {
// 	RefreshTokenID uuid.UUID `json:"refresh_token_id" db:"refresh_token_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
// 	UserID         uuid.UUID `json:"user_id" db:"user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
// 	TokenHash      string    `json:"token_hash" db:"token_hash" gorm:"type:varchar(255);not null"`
// 	ExpiresAt      time.Time `json:"expires_at" db:"expires_at" gorm:"not null"`
// 	IsRevoked      bool      `json:"is_revoked" db:"is_revoked" gorm:"default:false"`
// 	TokenCreatedAt time.Time `json:"token_created_at" db:"token_created_at" gorm:"default:CURRENT_TIMESTAMP"`

// 	// // Relationship
// 	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
// }

// func (UserRefreshToken) TableName() string {
// 	return "tbl_user_refresh_tokens"
// }

package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserToken represents a user token (refresh token or password reset token)
type UserToken struct {
	TokenID   uuid.UUID  `db:"token_id" json:"token_id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	TokenHash string     `db:"token_hash" json:"token_hash"`
	TokenType string     `db:"token_type" json:"token_type"` // "refresh" or "password_reset"
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	IsRevoked bool       `db:"is_revoked" json:"is_revoked"`
	IsUsed    bool       `db:"is_used" json:"is_used"`
	UsedAt    *time.Time `db:"used_at" json:"used_at,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

// TableName returns the table name for this entity
func (UserToken) TableName() string {
	return "tbl_user_tokens"
}
