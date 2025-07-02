package dtobookings

import "time"

type CreateBookingRequest struct {
	BookingID        string    `json:"booking_id"`
	UserID           string    `json:"user_id"`
	ExpertProfileID  string    `json:"expert_profile_id"`
	BookingDatetime  time.Time `json:"booking_datetime"`
	DurationMinutes  int       `json:"duration_minutes"`
	ConsultationType string    `json:"consultation_type"`
	BookingStatus    string    `json:"booking_status"`
	UserNotes        *string   `json:"user_notes,omitempty"`
	ExpertNotes      *string   `json:"expert_notes,omitempty"`
	MeetingLink      *string   `json:"meeting_link,omitempty"`
	MeetingAddress   *string   `json:"meeting_address,omitempty"`
	ConsultationFee  *float64  `json:"consultation_fee,omitempty"`
	PaymentStatus    string    `json:"payment_status"`
}

type CreateBookingResponse struct {
	BookingID        string    `json:"booking_id"`
	UserID           string    `json:"user_id"`
	ExpertProfileID  string    `json:"expert_profile_id"`
	BookingDatetime  time.Time `json:"booking_datetime"`
	DurationMinutes  int       `json:"duration_minutes"`
	ConsultationType string    `json:"consultation_type"`
	BookingStatus    string    `json:"booking_status"`
	UserNotes        *string   `json:"user_notes,omitempty"`
	ExpertNotes      *string   `json:"expert_notes,omitempty"`
	MeetingLink      *string   `json:"meeting_link,omitempty"`
	MeetingAddress   *string   `json:"meeting_address,omitempty"`
	ConsultationFee  *float64  `json:"consultation_fee,omitempty"`
	PaymentStatus    string    `json:"payment_status"`
	BookingCreatedAt time.Time `json:"booking_created_at"`
}

// CancellationReason *string    `json:"cancellation_reason,omitempty"`
// 	CancelledByUserID  *uuid.UUID `json:"cancelled_by_user_id,omitempty" db:"cancelled_by_user_id" gorm:"type:uuid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
// 	CancelledAt        *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
// 	ReminderSent       bool       `json:"reminder_sent" db:"reminder_sent" gorm:"default:false"`

// 	BookingUpdatedAt   time.Time  `json:"booking_updated_at" db:"booking_updated_at" gorm:"default:CURRENT_TIMESTAMP"`
