package entity

import (
	"time"

	"github.com/google/uuid"
)

// PricingConfig represents tbl_pricing_configs table
type PricingConfig struct {
	PricingID          uuid.UUID  `json:"pricing_id" db:"pricing_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ExpertProfileID    *uuid.UUID `json:"expert_profile_id,omitempty" db:"expert_profile_id" gorm:"type:uuid;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ServiceType        string     `json:"service_type" db:"service_type" gorm:"type:varchar(50);not null"`
	ConsultationType   string     `json:"consultation_type" db:"consultation_type" gorm:"type:varchar(20);not null;check:consultation_type IN ('online', 'offline')"`
	DurationMinutes    int        `json:"duration_minutes" db:"duration_minutes" gorm:"not null"`
	BasePrice          float64    `json:"base_price" db:"base_price" gorm:"type:decimal(10,2);not null"`
	DiscountPercentage float64    `json:"discount_percentage" db:"discount_percentage" gorm:"type:decimal(5,2);default:0"`
	IsActive           bool       `json:"is_active" db:"is_active" gorm:"default:true"`
	ValidFrom          time.Time  `json:"valid_from" db:"valid_from" gorm:"default:CURRENT_TIMESTAMP"`
	ValidUntil         *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	PricingCreatedAt   time.Time  `json:"pricing_created_at" db:"pricing_created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// // Relationship
	// ExpertProfile *entity.ExpertProfile `json:"expert_profile,omitempty" gorm:"foreignKey:ExpertProfileID;references:ExpertProfileID"`
}
