package dtobookings

import "time"

type GetAvailableSlotsRequest struct {
	ExpertProfileID     string    `json:"expert_profile_id" validate:"required,uuid"`
	FromDate            time.Time `json:"from_date" validate:"required"`
	ToDate              time.Time `json:"to_date" validate:"required"`
	SlotDurationMinutes int       `json:"slot_duration_minutes" validate:"min=15,max=240"` // Default 60 phút
}

type TimeSlot struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationMinutes int       `json:"duration_minutes"`
	IsAvailable     bool      `json:"is_available"`
}

type GetAvailableSlotsResponse struct {
	ExpertProfileID string     `json:"expert_profile_id"`
	FromDate        time.Time  `json:"from_date"`
	ToDate          time.Time  `json:"to_date"`
	AvailableSlots  []TimeSlot `json:"available_slots"`
	TotalSlots      int        `json:"total_slots"`
	Message         string     `json:"message,omitempty"`
}
