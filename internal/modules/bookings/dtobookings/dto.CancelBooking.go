package dtobookings

import "time"

type CancelBookingRequest struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
}

type CancelResponse struct {
	BookingID      string    `json:"booking_id"`
	CancelByUserID string    `json:"user_id"`
	Status         string    `json:"status"`
	CancelledAt    time.Time `json:"cancel_at"`
}
