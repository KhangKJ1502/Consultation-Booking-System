package dtoreviews

import "time"

type CreateReviewRequest struct {
	BookingID      string `json:"booking_id" validate:"required,uuid"`
	ReviewerUserID string `json:"reviewer_user_id" validate:"required,uuid"`
	RatingScore    int    `json:"rating_score" validate:"required,min=1,max=5"`
	ReviewComment  string `json:"review_comment" validate:"max=2000"`
	IsAnonymous    bool   `json:"is_anonymous"`
}

type CreateReviewResponse struct {
	ReviewID        string    `json:"review_id"`
	BookingID       string    `json:"booking_id"`
	ReviewerUserID  string    `json:"reviewer_user_id"`
	ExpertProfileID string    `json:"expert_profile_id"`
	RatingScore     int       `json:"rating_score"`
	ReviewComment   string    `json:"review_comment"`
	IsAnonymous     bool      `json:"is_anonymous"`
	ReviewCreatedAt time.Time `json:"review_created_at"`
}
