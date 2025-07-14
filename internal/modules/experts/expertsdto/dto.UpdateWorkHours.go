package dtoexperts

import (
	"time"
)

type UpdateWorkingHourRequest struct {
	WorkingHourID   string `json:"working_hour_id"`
	ExpertProfileID string `json:"expert_profile_id"`
	DayOfWeek       int    `json:"day_of_week"` // 0 = Chủ nhật, 1 = Thứ 2,...
	StartTime       string `json:"start_time"`  // "08:00"
	EndTime         string `json:"end_time"`    // "17:00"
}

type UpdateWorkingHourResponse struct {
	WorkingHourID string    `json:"working_hour_id"`
	DayOfWeek     int       `json:"day_of_week"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsActive      bool      `json:"is_active"`
}
