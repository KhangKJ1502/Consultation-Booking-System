package dtoexperts

import (
	"time"
)

type CreateWorkingHourRequest struct {
	ExpertProfileID string    `json:"expert_profile_id"`
	DayOfWeek       int       `json:"day_of_week"` // 0 = Chủ nhật, 1 = Thứ 2,...
	StartTime       time.Time `json:"start_time"`  // "08:00"
	EndTime         time.Time `json:"end_time"`    // "17:00"
}

type CreateWorkingHourResponse struct {
	WorkingHourID string    `json:"working_hour_id"`
	DayOfWeek     int       `json:"day_of_week"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsActive      bool      `json:"is_active"`
}
