package entity

import (
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	entityUser "cbs_backend/internal/modules/users/entity"
	"time"

	"github.com/google/uuid"
)

type ConsultationReview struct {
	ReviewID        uuid.UUID `json:"review_id" db:"review_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BookingID       uuid.UUID `json:"booking_id" db:"booking_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ReviewerUserID  uuid.UUID `json:"reviewer_user_id" db:"reviewer_user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ExpertProfileID uuid.UUID `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RatingScore     int       `json:"rating_score" db:"rating_score" gorm:"not null;check:rating_score BETWEEN 1 AND 5"`
	ReviewComment   *string   `json:"review_comment,omitempty" db:"review_comment" gorm:"type:text"`
	IsAnonymous     bool      `json:"is_anonymous" db:"is_anonymous" gorm:"default:false"`
	IsVisible       bool      `json:"is_visible" db:"is_visible" gorm:"default:true"`
	ReviewCreatedAt time.Time `json:"review_created_at" db:"review_created_at" gorm:"default:CURRENT_TIMESTAMP"`
	ReviewUpdatedAt time.Time `json:"review_updated_at" db:"review_updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// // Relationships (optional preload)
	Booking *entityBooking.ConsultationBooking `json:"booking,omitempty" gorm:"foreignKey:BookingID;references:BookingID"`

	ReviewerUser *entityUser.User `json:"reviewer_user,omitempty" gorm:"foreignKey:ReviewerUserID;references:UserID"`

	// ExpertProfile *entityExpert.ExpertProfile        `json:"expert_profile,omitempty" gorm:"foreignKey:ExpertProfileID;references:ExpertProfileID"`
}

func (ConsultationReview) TableName() string {
	return "tbl_consultation_reviews"
}
