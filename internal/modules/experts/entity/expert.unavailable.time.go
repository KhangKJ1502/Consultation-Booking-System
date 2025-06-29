package entity

import (
	"time"

	"cbs_backend/internal/common"

	"github.com/google/uuid"
)

// ExpertUnavailableTime represents tbl_expert_unavailable_times table
type ExpertUnavailableTime struct {
	UnavailableTimeID        uuid.UUID    `json:"unavailable_time_id" db:"unavailable_time_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ExpertProfileID          uuid.UUID    `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UnavailableStartDatetime time.Time    `json:"unavailable_start_datetime" db:"unavailable_start_datetime" gorm:"not null"`
	UnavailableEndDatetime   time.Time    `json:"unavailable_end_datetime" db:"unavailable_end_datetime" gorm:"not null"`
	UnavailableReason        *string      `json:"unavailable_reason,omitempty" db:"unavailable_reason" gorm:"type:text"`
	IsRecurring              bool         `json:"is_recurring" db:"is_recurring" gorm:"default:false"`
	RecurrencePattern        common.JSONB `json:"recurrence_pattern,omitempty" db:"recurrence_pattern" gorm:"type:jsonb"`
	UnavailableCreatedAt     time.Time    `json:"unavailable_created_at" db:"unavailable_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// // Relationships
	ExpertProfile ExpertProfile `json:"expert_profile" gorm:"foreignKey:ExpertProfileID;references:ExpertProfileID"`
}

func (ExpertUnavailableTime) TableName() string {
	return "tbl_expert_unavailable_times"
}
