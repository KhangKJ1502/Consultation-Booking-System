package dtoexperts

import (
	"time"
)

type WorkingHourDTO struct {
	DayOfWeek string `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type UnavailableTimeDTO struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type ExpertFullDetailResponse struct {
	ExpertProfileID    string               `json:"expert_profile_id"`
	SpecializationList []string             `json:"specialization_list"`
	ExperienceYears    int                  `json:"experience_years"`
	ExpertBio          string               `json:"expert_bio"`
	ConsultationFee    float64              `json:"consultation_fee"`
	AverageRating      float64              `json:"average_rating"`
	TotalReviews       int                  `json:"total_reviews"`
	IsVerified         bool                 `json:"is_verified"`
	LicenseNumber      string               `json:"license_number"`
	AvailableOnline    bool                 `json:"available_online"`
	AvailableOffline   bool                 `json:"available_offline"`
	User               UserDTO              `json:"user"`
	WorkingHours       []WorkingHourDTO     `json:"working_hours"`
	UnavailableTimes   []UnavailableTimeDTO `json:"unavailable_times"`
}
