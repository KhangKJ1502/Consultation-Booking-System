package dtodashboard

type SystemOverviewResponse struct {
	TotalBookings     int64   `json:"total_bookings"`
	PendingBookings   int64   `json:"pending_bookings"`
	ConfirmedBookings int64   `json:"confirmed_bookings"`
	CompletedBookings int64   `json:"completed_bookings"`
	CancelledBookings int64   `json:"cancelled_bookings"`
	ActiveExperts     int64   `json:"active_experts"`
	ActiveUsers       int64   `json:"active_users"`
	TotalRevenue      float64 `json:"total_revenue"`
	SuccessRate       float64 `json:"success_rate"`
}
