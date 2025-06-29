package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserRefreshToken represents tbl_user_refresh_tokens table
type UserRefreshToken struct {
	RefreshTokenID uuid.UUID `json:"refresh_token_id" db:"refresh_token_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID         uuid.UUID `json:"user_id" db:"user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TokenHash      string    `json:"token_hash" db:"token_hash" gorm:"type:varchar(255);not null"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at" gorm:"not null"`
	IsRevoked      bool      `json:"is_revoked" db:"is_revoked" gorm:"default:false"`
	TokenCreatedAt time.Time `json:"token_created_at" db:"token_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// // Relationship
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}

func (UserRefreshToken) TableName() string {
	return "tbl_user_refresh_tokens"
}
