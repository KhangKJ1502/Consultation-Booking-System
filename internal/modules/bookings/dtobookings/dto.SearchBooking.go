package dtobookings

import "time"

type SearchBookingsRequest struct {
	UserID           string    `json:"user_id,omitempty"`
	ExpertProfileID  string    `json:"expert_profile_id,omitempty"`
	Status           string    `json:"status,omitempty"`
	ConsultationType string    `json:"consultation_type,omitempty"`
	FromDate         time.Time `json:"from_date,omitempty"`
	ToDate           time.Time `json:"to_date,omitempty"`
	Page             int       `json:"page" validate:"min=1"`
	PageSize         int       `json:"page_size" validate:"min=1,max=100"`
}

type SearchBookingsResponse struct {
	Results     []BookingResponse `json:"results"`
	TotalCount  int               `json:"total_count"`
	CurrentPage int               `json:"current_page"`
	PageSize    int               `json:"page_size"`
	TotalPages  int               `json:"total_pages"`
}
