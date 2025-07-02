package dtobookings

import "time"

type BookingResponse struct {
	BookingID        string    `json:"booking_id"`
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
