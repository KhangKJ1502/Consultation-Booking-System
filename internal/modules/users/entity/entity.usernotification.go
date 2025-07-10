package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserNotificationSetting struct {
	SettingID        uuid.UUID `json:"setting_id" db:"setting_id" gorm:"primaryKey;default:uuid_generate_v4()"`
	UserID           uuid.UUID `json:"user_id" db:"user_id" gorm:"not null;constraint:OnDelete:CASCADE"`
	NotificationType string    `json:"notification_type" db:"notification_type" gorm:"not null"`
	DeliveryMethod   string    `json:"delivery_method" db:"delivery_method" gorm:"not null"`
	IsEnabled        bool      `json:"is_enabled" db:"is_enabled" gorm:"default:true"`
	CreatedAt        time.Time `json:"created_at" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at" gorm:"autoUpdateTime"`
}

func (UserNotificationSetting) TableName() string {
	return "tbl_user_notification_settings"
}
