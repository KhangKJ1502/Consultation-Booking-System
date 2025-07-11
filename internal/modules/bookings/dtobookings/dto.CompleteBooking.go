package dtobookings

import "time"

type CompleteBookingRequest struct {
	BookingID string `json:"booking_id" validate:"required"`
	ExpertID  string `json:"expert_id" validate:"required"`
}

type CompleteBookingResponse struct {
	BookingID   string    `json:"booking_id"`
	Status      string    `json:"status"`
	CompletedAt time.Time `json:"completed_at"`
	Message     string    `json:"message"`
}
