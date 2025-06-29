package entity

import (
	"time"

	"github.com/google/uuid"
)

// ExpertSpecialization represents tbl_expert_specializations table // chuyên môn chuyên môn
type ExpertSpecialization struct {
	SpecializationID          uuid.UUID `json:"specialization_id" db:"specialization_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ExpertProfileID           uuid.UUID `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SpecializationName        string    `json:"specialization_name" db:"specialization_name" gorm:"type:varchar(100);not null"`
	SpecializationDescription *string   `json:"specialization_description,omitempty" db:"specialization_description" gorm:"type:text"`
	IsPrimary                 bool      `json:"is_primary" db:"is_primary" gorm:"default:false"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	ExpertProfile *ExpertProfile `json:"expert_profile,omitempty" gorm:"foreignKey:ExpertProfileID;references:ExpertProfileID"`
}

func (ExpertSpecialization) TableName() string {
	return "tbl_expert_specializations"
}
