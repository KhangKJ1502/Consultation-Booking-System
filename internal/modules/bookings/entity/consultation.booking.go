package entity

import (
	"time"

	entityExpertProfile "cbs_backend/internal/modules/experts/entity"
	entityUsers "cbs_backend/internal/modules/users/entity"

	"github.com/google/uuid"
)

// ConsultationBooking represents tbl_consultation_bookings table
type ConsultationBooking struct {
	BookingID          uuid.UUID  `json:"booking_id" db:"booking_id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ExpertProfileID    uuid.UUID  `json:"expert_profile_id" db:"expert_profile_id" gorm:"type:uuid;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingDatetime    time.Time  `json:"booking_datetime" db:"booking_datetime" gorm:"not null"`
	DurationMinutes    int        `json:"duration_minutes" db:"duration_minutes" gorm:"default:60"`
	ConsultationType   string     `json:"consultation_type" db:"consultation_type" gorm:"type:varchar(20);not null;check:consultation_type IN ('online', 'offline')"`
	BookingStatus      string     `json:"booking_status" db:"booking_status" gorm:"type:varchar(20);not null;default:'pending';check:booking_status IN ('pending', 'confirmed', 'rejected', 'cancelled', 'completed', 'missed', 'no_show')"`
	UserNotes          *string    `json:"user_notes,omitempty" db:"user_notes" gorm:"type:text"`
	ExpertNotes        *string    `json:"expert_notes,omitempty" db:"expert_notes" gorm:"type:text"`
	MeetingLink        *string    `json:"meeting_link,omitempty" db:"meeting_link" gorm:"type:text"`
	MeetingAddress     *string    `json:"meeting_address,omitempty" db:"meeting_address" gorm:"type:text"`
	ConsultationFee    *float64   `json:"consultation_fee,omitempty" db:"consultation_fee" gorm:"type:decimal(10,2)"`
	PaymentStatus      string     `json:"payment_status" db:"payment_status" gorm:"type:varchar(20);default:'pending';check:payment_status IN ('pending', 'paid', 'refunded', 'failed');constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CancellationReason *string    `json:"cancellation_reason,omitempty" db:"cancellation_reason" gorm:"type:text"`
	CancelledByUserID  *uuid.UUID `json:"cancelled_by_user_id,omitempty" db:"cancelled_by_user_id" gorm:"type:uuid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	ReminderSent       bool       `json:"reminder_sent" db:"reminder_sent" gorm:"default:false"`
	BookingCreatedAt   time.Time  `json:"booking_created_at" db:"booking_created_at" gorm:"default:CURRENT_TIMESTAMP"`
	BookingUpdatedAt   time.Time  `json:"booking_updated_at" db:"booking_updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	User            entityUsers.User                  `json:"user" gorm:"foreignKey:UserID"`
	ExpertProfile   entityExpertProfile.ExpertProfile `json:"expert_profile" gorm:"foreignKey:ExpertProfileID"`
	CancelledByUser *entityUsers.User                 `json:"cancelled_by_user,omitempty" gorm:"foreignKey:CancelledByUserID"`
	// StatusHistory []BookingStatusHistory `json:"status_history,omitempty" gorm:"foreignKey:BookingID"`
	// Review              *ConsultationReview                `json:"review,omitempty" gorm:"foreignKey:BookingID"`
	// PaymentTransactions []entityPayment.PaymentTransaction `json:"payment_transactions,omitempty" gorm:"foreignKey:BookingID"`
}
