package dtodashboard

type ExpertPerformanceResponse struct {
	ExpertID          string  `json:"expert_id"`
	ExpertName        string  `json:"expert_name"`
	TotalBookings     int64   `json:"total_bookings"`
	CompletedBookings int64   `json:"completed_bookings"`
	CancelledBookings int64   `json:"cancelled_bookings"`
	Revenue           float64 `json:"revenue"`
	AverageRating     float64 `json:"average_rating"`
	SuccessRate       float64 `json:"success_rate"`
}
