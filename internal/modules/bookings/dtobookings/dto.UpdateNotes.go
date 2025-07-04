package dtobookings

import "time"

// 2. DTOs cho UpdateBookingNotes
type UpdateBookingNotesRequest struct {
	BookingID string `json:"booking_id" validate:"required,uuid"`
	UserID    string `json:"user_id" validate:"required,uuid"`
	Notes     string `json:"notes" validate:"max=1000"` // Giới hạn 1000 ký tự
}

type UpdateBookingNotesResponse struct {
	BookingID string    `json:"booking_id"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"` // "user" or "expert"
	Message   string    `json:"message"`
}
