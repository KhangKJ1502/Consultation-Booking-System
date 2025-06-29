package dtoexperts

import (
// "github.com/lib/pq"
)

type GetAllExpertsRespone struct {
	ExpertProfileID    string   `json:"expert_profile_id"`
	SpecializationList []string `json:"specialization_list"`
	ExperienceYears    *int     `json:"experience_years,omitempty"`
	ConsultationFee    *float64 `json:"consultation_fee,omitempty"`
	AverageRating      float64  `json:"average_rating"`
	TotalReviews       int      `json:"total_reviews"`
	User               UserDTO  `json:"user"`
}
