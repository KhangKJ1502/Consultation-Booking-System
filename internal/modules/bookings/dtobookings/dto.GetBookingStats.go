package dtobookings

import "time"

type GetBookingStatsRequest struct {
	UserID   string    `json:"user_id" validate:"required"`
	FromDate time.Time `json:"from_date" validate:"required"`
	ToDate   time.Time `json:"to_date" validate:"required"`
}

type BookingStats struct {
	TotalBookings int64            `json:"total_bookings"`
	StatusCounts  map[string]int64 `json:"status_counts"`
	TotalSpent    float64          `json:"total_spent"`
}

type GetBookingStatsResponse struct {
	UserID      string       `json:"user_id"`
	FromDate    time.Time    `json:"from_date"`
	ToDate      time.Time    `json:"to_date"`
	Stats       BookingStats `json:"stats"`
	GeneratedAt time.Time    `json:"generated_at"`
}
