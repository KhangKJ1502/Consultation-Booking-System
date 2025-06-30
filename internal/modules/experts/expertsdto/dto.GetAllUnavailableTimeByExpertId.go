package dtoexperts

import "time"

type GetAllsExpertUnavailableTimeResponse struct {
	UnavailableTimeID string    `json:"unavailable_time_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	Reason            *string   `json:"reason,omitempty"`
	IsRecurring       bool      `json:"is_recurring"`
	RecurrencePattern *string   `json:"recurrence_pattern,omitempty"`
}
