package dtobookings

import "time"

type GetUserBookingHistoryRequest struct {
	UserID   string    `json:"user_id" validate:"required"`
	Status   string    `json:"status,omitempty"`
	FromDate time.Time `json:"from_date,omitempty"`
	ToDate   time.Time `json:"to_date,omitempty"`
	Page     int       `json:"page" validate:"min=1"`
	PageSize int       `json:"page_size" validate:"min=1,max=100"`
}

type GetUserBookingHistoryResponse struct {
	UserID      string            `json:"user_id"`
	Bookings    []BookingResponse `json:"bookings"`
	TotalCount  int               `json:"total_count"`
	CurrentPage int               `json:"current_page"`
	PageSize    int               `json:"page_size"`
	TotalPages  int               `json:"total_pages"`
}
