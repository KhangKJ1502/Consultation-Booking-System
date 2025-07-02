package dtoexperts

type UpdateExpertSpecializationRequest struct {
	SpecializationID          string `json:"specialization_id" binding:"required"`
	ExpertProfileID           string `json:"expert_profile_id" binding:"required"`
	SpecializationName        string `json:"specialization_name" binding:"required"`
	SpecializationDescription string `json:"specialization_description" binding:"required"`
	IsPrimary                 bool   `json:"is_primary"`
}

type UpdateExpertSpecializationRespone struct {
	SpecializationID          string `json:"expert_profile_id" binding:"required"`
	ExpertProfileID           string `json:"expert_profile_id"`
	SpecializationName        string `json:"specialization_name"`
	SpecializationDescription string `json:"specialization_description"`
	IsPrimary                 bool   `json:"is_primary"`
}
