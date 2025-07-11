package dtobookings

import "time"

type RescheduleBookingRequest struct {
	BookingID          string    `json:"booking_id" validate:"required"`
	UserID             string    `json:"user_id" validate:"required"`
	ExpertProfileID    string    `json:"expert_profile_id" validate:"required"`
	NewBookingDatetime time.Time `json:"new_booking_datetime" validate:"required"`
	RescheduleReason   string    `json:"reschedule_reason"`
}

type RescheduleBookingResponse struct {
	BookingID          string    `json:"booking_id"`
	OldBookingDatetime time.Time `json:"old_booking_datetime"`
	NewBookingDatetime time.Time `json:"new_booking_datetime"`
	BookingStatus      string    `json:"booking_status"`
	RescheduledAt      time.Time `json:"rescheduled_at"`
	Message            string    `json:"message"`
}
