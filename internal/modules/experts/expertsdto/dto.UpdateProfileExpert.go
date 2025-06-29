package dtoexperts

import (
	"time"
)

type UpdateProfileExpertRequest struct {
	ExpertProfileID    string   `json:"expert_profile_id" validate:"required"`
	SpecializationList []string `json:"specialization_list" validate:"required"`
	ExperienceYears    *int     `json:"experience_years,omitempty"`
	ExpertBio          *string  `json:"expert_bio,omitempty"`
	ConsultationFee    *float64 `json:"consultation_fee,omitempty"`
	LicenseNumber      *string  `json:"license_number,omitempty"`
	AvailableOnline    bool     `json:"available_online"`
	AvailableOffline   bool     `json:"available_offline"`
}

type UpdateProfileExpertResponse struct {
	ExpertProfileID    string    `json:"expert_profile_id"`
	SpecializationList []string  `json:"specialization_list"`
	ExperienceYears    *int      `json:"experience_years,omitempty"`
	ExpertBio          *string   `json:"expert_bio,omitempty"`
	ConsultationFee    *float64  `json:"consultation_fee,omitempty"`
	AverageRating      float64   `json:"average_rating"`
	TotalReviews       int       `json:"total_reviews"`
	IsVerified         bool      `json:"is_verified"`
	LicenseNumber      *string   `json:"license_number,omitempty"`
	AvailableOnline    bool      `json:"available_online"`
	AvailableOffline   bool      `json:"available_offline"`
	ExpertCreatedAt    time.Time `json:"expert_created_at"`
	ExpertUpdatedAt    time.Time `json:"expert_updated_at"`
	User               UserDTO   `json:"user"`
}
