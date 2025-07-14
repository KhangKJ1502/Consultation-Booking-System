package dtoexperts

import (
	"time"
)

type UpdateUnavailableTimeRequest struct {
	UnavailableTimeID string  `json:"unavailable_time_id" binding:"required"`
	ExpertProfileID   string  `json:"expert_profile_id" binding:"required"`
	StartDatetime     string  `json:"start_datetime" binding:"required"` // đổi sang string
	EndDatetime       string  `json:"end_datetime" binding:"required"`
	Reason            *string `json:"reason"`
	IsRecurring       bool    `json:"is_recurring"`
	RecurrencePattern any     `json:"recurrence_pattern"` // nếu có dùng
}

type UpdateUnavailableTimeResponse struct {
	UnavailableTimeID string    `json:"unavailable_time_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	Reason            *string   `json:"reason"`
	IsRecurring       bool      `json:"is_recurring"`
	RecurrencePattern any       `json:"recurrence_pattern"`
}
