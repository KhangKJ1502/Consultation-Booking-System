package dtodashboard

import "time"

type RevenueReportRequest struct {
	DateFrom time.Time `json:"date_from"`
	DateTo   time.Time `json:"date_to"`
	GroupBy  string    `json:"group_by"` // day, week, month
}

type RevenueReportResponse struct {
	Period       string  `json:"period"`
	Revenue      float64 `json:"revenue"`
	BookingCount int64   `json:"booking_count"`
	Growth       float64 `json:"growth_percentage"`
}
