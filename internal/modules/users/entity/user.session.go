package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserSession represents tbl_user_sessions table
type UserSession struct {
	SessionID        uuid.UUID `json:"session_id" db:"session_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID           uuid.UUID `json:"user_id" db:"user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SessionToken     string    `json:"session_token" db:"session_token" gorm:"type:varchar(255);not null"`
	DeviceInfo       *string   `json:"device_info,omitempty" db:"device_info" gorm:"type:text"`
	IPAddress        *string   `json:"ip_address,omitempty" db:"ip_address" gorm:"type:inet"`
	UserAgent        *string   `json:"user_agent,omitempty" db:"user_agent" gorm:"type:text"`
	IsActive         bool      `json:"is_active" db:"is_active" gorm:"default:true"`
	LastActivity     time.Time `json:"last_activity" db:"last_activity" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt        time.Time `json:"expires_at" db:"expires_at" gorm:"not null"`
	SessionCreatedAt time.Time `json:"session_created_at" db:"session_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// // Relationship
	// User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:UserID"`
}
