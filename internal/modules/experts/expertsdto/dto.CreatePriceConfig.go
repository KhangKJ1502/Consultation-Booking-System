package dtoexperts

import "time"

type CreatePricingConfigRequest struct {
	ExpertProfileID    string     `json:"expert_profile_id" binding:"required"`
	ServiceType        string     `json:"service_type" binding:"required"`
	ConsultationType   string     `json:"consultation_type" binding:"required"`
	DurationMinutes    int        `json:"duration_minutes" binding:"required"`
	BasePrice          float64    `json:"base_price" binding:"required"`
	DiscountPercentage float64    `json:"discount_percentage"`
	IsActive           bool       `json:"is_active" binding:"required"`
	ValidFrom          time.Time  `json:"valid_from" binding:"required"`
	ValidUntil         *time.Time `json:"valid_until,omitempty" binding:"required"`
}
type CreatePricingConfigResponse struct {
	PricingID          string     `json:"pricing_id"`
	ExpertProfileID    string     `json:"expert_profile_id" binding:"required"`
	ServiceType        string     `json:"service_type" binding:"required"`
	ConsultationType   string     `json:"consultation_type"`
	DurationMinutes    int        `json:"duration_minutes"`
	BasePrice          float64    `json:"base_price"`
	DiscountPercentage float64    `json:"discount_percentage"`
	IsActive           bool       `json:"is_active"`
	ValidFrom          time.Time  `json:"valid_from"`
	ValidUntil         *time.Time `json:"valid_until,omitempty"`
	PricingCreatedAt   time.Time  `json:"pricing_created_at"`
}
