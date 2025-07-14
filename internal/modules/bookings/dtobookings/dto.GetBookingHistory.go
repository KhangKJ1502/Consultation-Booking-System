package dtobookings

import "time"

type GetBookingStatusHistoryRequest struct {
	BookingID string `json:"booking_id" binding:"required,uuid"`
	UserID    string `json:"user_id" binding:"required,uuid"`
}

type StatusHistoryItem struct {
	StatusHistoryID string    `json:"status_history_id"`
	OldStatus       string    `json:"old_status"`
	NewStatus       string    `json:"new_status"`
	ChangedByUserID string    `json:"changed_by_user_id"`
	ChangedByName   string    `json:"changed_by_name"`
	ChangeReason    string    `json:"change_reason"`
	StatusChangedAt time.Time `json:"status_changed_at"`
}
type GetBookingStatusHistoryResponse struct {
	BookingID     string              `json:"booking_id"`
	StatusHistory []StatusHistoryItem `json:"status_history"`
	TotalRecords  int                 `json:"total_records"`
}
