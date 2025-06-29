package entity

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatusHistory struct {
	StatusHistoryID uuid.UUID `json:"status_history_id" db:"status_history_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BookingID       uuid.UUID `json:"booking_id" db:"booking_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	OldStatus       *string   `json:"old_status,omitempty" db:"old_status" gorm:"type:varchar(20)"`
	NewStatus       string    `json:"new_status" db:"new_status" gorm:"type:varchar(20);not null"`
	ChangedByUserID uuid.UUID `json:"changed_by_user_id" db:"changed_by_user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ChangeReason    *string   `json:"change_reason,omitempty" db:"change_reason" gorm:"type:text"`
	StatusChangedAt time.Time `json:"status_changed_at" db:"status_changed_at" gorm:"default:CURRENT_TIMESTAMP"`
}
