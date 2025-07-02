package dtobookings

// 2. DTOs cho UpdateBookingNotes
type UpdateBookingNotesRequest struct {
	BookingID string `json:"booking_id" validate:"required,uuid"`
	UserID    string `json:"user_id" validate:"required,uuid"`
	Notes     string `json:"notes" validate:"max=1000"` // Giới hạn 1000 ký tự
}
