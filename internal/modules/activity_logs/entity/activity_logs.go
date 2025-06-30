package entityactivitylog

import (
	"time"

	"cbs_backend/internal/common"
	"cbs_backend/internal/modules/users/entity"

	"github.com/google/uuid"
)

type ActivityLog struct {
	LogID            uuid.UUID    `json:"log_id" db:"log_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID           *uuid.UUID   `json:"user_id,omitempty" db:"user_id" gorm:"type:uuid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ActionPerformed  string       `json:"action_performed" db:"action_performed" gorm:"type:varchar(100);not null"`
	AffectedTable    string       `json:"affected_table" db:"affected_table" gorm:"type:varchar(50);not null"`
	AffectedRecordID *uuid.UUID   `json:"affected_record_id,omitempty" db:"affected_record_id" gorm:"type:uuid"`
	OldValues        common.JSONB `json:"old_values,omitempty" db:"old_values" gorm:"type:jsonb"`
	NewValues        common.JSONB `json:"new_values,omitempty" db:"new_values" gorm:"type:jsonb"`
	UserIPAddress    *string      `json:"user_ip_address,omitempty" db:"user_ip_address" gorm:"type:inet"`
	UserAgent        *string      `json:"user_agent,omitempty" db:"user_agent" gorm:"type:text"`
	LogCreatedAt     time.Time    `json:"log_created_at" db:"log_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	User *entity.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

func (ActivityLog) TableName() string {
	return "tbl_activity_logs"
}
