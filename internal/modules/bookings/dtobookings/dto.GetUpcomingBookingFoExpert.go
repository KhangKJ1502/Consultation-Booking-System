package dtobookings

import "time"

type GetUpcomingBookingForExpertRequest struct {
	ExpertID string    `json:"expert_id"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
}
