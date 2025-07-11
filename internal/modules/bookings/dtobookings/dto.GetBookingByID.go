package dtobookings

import "time"

type GetBookingByIDRequest struct {
	BookingID string `json:"booking_id" validate:"required"`
	UserID    string `json:"user_id" validate:"required"`
}

type BookingDetailResponse struct {
	BookingID        string    `json:"booking_id"`
	UserID           string    `json:"user_id"`
	ExpertProfileID  string    `json:"expert_profile_id"`
	BookingDatetime  time.Time `json:"booking_datetime"`
	DurationMinutes  int       `json:"duration_minutes"`
	ConsultationType string    `json:"consultation_type"`
	BookingStatus    string    `json:"booking_status"`
	PaymentStatus    string    `json:"payment_status"`
	UserNotes        *string   `json:"user_notes"`
	ExpertNotes      *string   `json:"expert_notes"`
	MeetingLink      *string   `json:"meeting_link"`
	MeetingAddress   *string   `json:"meeting_address"`
	ConsultationFee  float64   `json:"consultation_fee"`
	BookingCreatedAt time.Time `json:"booking_created_at"`
	BookingUpdatedAt time.Time `json:"booking_updated_at"`
}
