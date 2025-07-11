package entity

import (
	"time"

	"cbs_backend/internal/common"
	"cbs_backend/internal/modules/users/entity"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SystemNotification represents tbl_system_notifications table
type SystemNotification struct {
	NotificationID        uuid.UUID      `json:"notification_id" db:"notification_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RecipientUserID       uuid.UUID      `json:"recipient_user_id" db:"recipient_user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	NotificationType      string         `json:"notification_type" db:"notification_type" gorm:"type:varchar(50);not null"`
	NotificationTitle     string         `json:"notification_title" db:"notification_title" gorm:"type:varchar(255);not null"`
	NotificationMessage   string         `json:"notification_message" db:"notification_message" gorm:"type:text;not null"`
	NotificationStatus    string         `json:"notification_status" db:"notification_status" gorm:"type:varchar(50);default:'pending'"`
	NotificationData      common.JSONB   `json:"notification_data,omitempty" db:"notification_data" gorm:"type:jsonb"`
	IsRead                bool           `json:"is_read" db:"is_read" gorm:"default:false"`
	DeliveryMethods       pq.StringArray `json:"delivery_methods" db:"delivery_methods" gorm:"type:varchar(20)[];default:ARRAY['app']"`
	SentAt                *time.Time     `json:"sent_at,omitempty" db:"sent_at"`
	ReadAt                *time.Time     `json:"read_at,omitempty" db:"read_at"`
	ExpiresAt             *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
	NotificationCreatedAt time.Time      `json:"notification_created_at" db:"notification_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	RecipientUser *entity.User `json:"recipient_user,omitempty" gorm:"foreignKey:RecipientUserID;references:UserID"`
}

func (SystemNotification) TableName() string {
	return "tbl_system_notifications"
}
