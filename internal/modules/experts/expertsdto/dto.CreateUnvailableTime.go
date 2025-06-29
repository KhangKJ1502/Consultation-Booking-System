package dtoexperts

import (
	"time"
)

type CreateUnavailableTimeRequest struct {
	ExpertProfileID   string    `json:"expert_profile_id" binding:"required"`
	StartDatetime     time.Time `json:"start_time" binding:"required"`
	EndDatetime       time.Time `json:"end_time" binding:"required"`
	Reason            *string   `json:"reason"`
	IsRecurring       bool      `json:"is_recurring"`
	RecurrencePattern any       `json:"recurrence_pattern"` // JSON dạng rules
}

type CreateUnavailableTimeResponse struct {
	UnavailableTimeID string    `json:"unavailable_time_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	Reason            *string   `json:"reason"`
	IsRecurring       bool      `json:"is_recurring"`
	RecurrencePattern any       `json:"recurrence_pattern"`
}
