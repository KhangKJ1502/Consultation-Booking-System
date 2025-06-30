package dtoexperts

import "time"

type GetAllExpertSpecializationRespone struct {
	SpecializationID          string    `json:"specialization_id"`
	ExpertProfileID           string    `json:"expert_profile_id"`
	SpecializationName        string    `json:"specialization_name" `
	SpecializationDescription string    `json:"specialization_description"`
	IsPrimary                 bool      `json:"is_primary"`
	CreateAt                  time.Time `json:"create_at"`
}
