package dtobookings

type CancelBookingRequest struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
}
