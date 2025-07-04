package dtobookings

import "time"

type ConfirmBooking struct {
	BookingID string `json:"booking_id" binding:"required"`
	ExpertID  string `json:"expert_id" binding:"required"`
}

type ConfirmBookingResponse struct {
	BookingID       string
	ExpertID        string
	UserID          string
	Status          string
	ConfirmAt       time.Time
	DurationMinutes int
	MeetingLink     string
	MeetingAddress  string
}
