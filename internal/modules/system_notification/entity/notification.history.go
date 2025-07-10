package entity

import (
	"cbs_backend/internal/modules/users/entity"
	"time"

	"github.com/google/uuid"
)

type NotificationHistory struct {
	NotificationID   uuid.UUID   `json:"notification_id" db:"notification_id" gorm:"primaryKey;default:uuid_generate_v4()"`
	UserID           uuid.UUID   `json:"user_id" db:"user_id" gorm:"not null;constraint:OnDelete:CASCADE"`
	User             entity.User `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	NotificationType string      `json:"notification_type" db:"notification_type" gorm:"not null"`
	DeliveryMethod   string      `json:"delivery_method" db:"delivery_method" gorm:"not null"`
	Title            string      `json:"title" db:"title" gorm:"not null"`
	Content          string      `json:"content" db:"content" gorm:"not null"`
	DeliveryStatus   string      `json:"delivery_status" db:"delivery_status" gorm:"not null;default:pending"`
	SentAt           *time.Time  `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt      *time.Time  `json:"delivered_at,omitempty" db:"delivered_at"`
	ErrorMessage     *string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at" gorm:"autoCreateTime"`
}

func (NotificationHistory) TableName() string {
	return "tbl_notification_history"
}
