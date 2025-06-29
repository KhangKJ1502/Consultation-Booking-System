package entity

import (
	"time"

	"github.com/google/uuid"
)

// ExpertWorkingHour represents tbl_expert_working_hours table
type ExpertWorkingHour struct {
	WorkingHourID        uuid.UUID `json:"working_hour_id" db:"working_hour_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ExpertProfileID      uuid.UUID `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ExpertProfileID;references:ExpertProfileID"`
	DayOfWeek            int       `json:"day_of_week" db:"day_of_week" gorm:"not null;check:day_of_week BETWEEN 0 AND 6"`
	StartTime            time.Time `json:"start_time" db:"start_time" gorm:"type:time;not null"`
	EndTime              time.Time `json:"end_time" db:"end_time" gorm:"type:time;not null"`
	IsActive             bool      `json:"is_active" db:"is_active" gorm:"default:true"`
	WorkingHourCreatedAt time.Time `json:"working_hour_created_at" db:"working_hour_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	ExpertProfile ExpertProfile `json:"expert_profile" gorm:"foreignKey:ExpertProfileID;references:ExpertProfileID"`
}

func (ExpertWorkingHour) TableName() string {
	return "tbl_expert_working_hours"
}
