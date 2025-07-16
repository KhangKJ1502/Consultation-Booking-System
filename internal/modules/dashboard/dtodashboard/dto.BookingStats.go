package dtodashboard

import "time"

type BookingStatsRequest struct {
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
	ExpertID *string    `json:"expert_id"`
	Status   *string    `json:"status"`
}

type BookingStatsResponse struct {
	Period  string  `json:"period"`
	Count   int64   `json:"count"`
	Revenue float64 `json:"revenue"`
	Status  string  `json:"status"`
}
