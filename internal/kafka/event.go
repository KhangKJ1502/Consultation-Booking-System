package kafka

import (
	"time"
)

// User Events
type UserRegisteredEvent struct {
	EventType    string    `json:"event_type"` // "user_registered"
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	RegisteredAt time.Time `json:"registered_at"`
}

type UserProfileUpdatedEvent struct {
	EventType string    `json:"event_type"` // "user_profile_updated"
	UserID    string    `json:"user_id"`
	Changes   []string  `json:"changes"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Notification Events
type NotificationEvent struct {
	UserID        string `json:"user_id"`
	RecipientID   string `json:"recipient_id"`   // Người nhận cụ thể
	RecipientType string `json:"recipient_type"` // "user", "expert", "admin"...

	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// ---------------------- BOOKING EVENT-------------------------
type BookingEvent struct {
	EventType string                 `json:"event_type"`
	BookingID string                 `json:"booking_id"`
	UserID    string                 `json:"user_id"`
	ExpertID  string                 `json:"expert_id"`
	Timestamp time.Time              `json:"timestamp"`
	EventData map[string]interface{} `json:"event_data"`
}

// Event structs - Định nghĩa các struct để parse events
type BookingCreatedEvent struct {
	EventType          string    `json:"event_type"`
	UserID             string    `json:"user_id"`
	BookingID          string    `json:"booking_id"`
	ExpertID           string    `json:"expert_id"`
	DoctorName         string    `json:"doctor_name"`
	DoctorSpecialty    []string  `json:"doctor_specialty"`
	ConsultationDate   string    `json:"consultation_date"`
	ConsultationTime   string    `json:"consultation_time"`
	Duration           int       `json:"duration"`
	ConsultationType   string    `json:"consultation_type"`
	Location           string    `json:"location"`
	MeetingLink        string    `json:"meeting_link"`
	Amount             float64   `json:"amount"`
	PaymentStatus      string    `json:"payment_status"`
	BookingNotes       string    `json:"booking_notes"`
	CancellationPolicy string    `json:"cancellation_policy"`
	Email              string    `json:"email"`
	FullName           string    `json:"full_name"`
	ConfirmedAt        time.Time `json:"confirmed_at"`
}

// BookingConfirmEvent
type BookingConfirmEvent struct {
	EventType          string    `json:"event_type"`
	UserID             string    `json:"user_id"`
	BookingID          string    `json:"booking_id"`
	ExpertID           string    `json:"expert_id"`
	DoctorName         string    `json:"doctor_name"`
	DoctorSpecialty    []string  `json:"doctor_specialty"`
	ConsultationDate   string    `json:"consultation_date"`
	ConsultationTime   string    `json:"consultation_time"`
	Duration           int       `json:"duration"`
	ConsultationType   string    `json:"consultation_type"`
	Location           string    `json:"location"`
	MeetingLink        string    `json:"meeting_link"`
	Amount             float64   `json:"amount"`
	PaymentStatus      string    `json:"payment_status"`
	BookingNotes       string    `json:"booking_notes"`
	CancellationPolicy string    `json:"cancellation_policy"`
	Email              string    `json:"email"`
	FullName           string    `json:"full_name"`
	ConfirmedAt        time.Time `json:"confirmed_at"`
}

// Booking Cancel Event
type BookingCancelledEvent struct {
	EventType         string    `json:"event_type"`          // e.g. "booking_cancelled"
	BookingID         string    `json:"booking_id"`          // ID của lịch hẹn bị huỷ
	UserID            string    `json:"user_id"`             // ID người dùng đặt lịch
	ExpertID          string    `json:"expert_id"`           // ID chuyên gia
	DoctorName        string    `json:"doctor_name"`         // Tên chuyên gia
	DoctorSpecialty   []string  `json:"doctor_specialty"`    // Chuyên khoa
	ConsultationDate  string    `json:"consultation_date"`   // Ngày tư vấn
	ConsultationTime  string    `json:"consultation_time"`   // Giờ tư vấn
	Duration          int       `json:"duration"`            // Thời lượng tư vấn (phút)
	ConsultationType  string    `json:"consultation_type"`   // Loại tư vấn
	Location          string    `json:"location"`            // Địa điểm (hoặc "online")
	MeetingLink       string    `json:"meeting_link"`        // Link cuộc hẹn (nếu online)
	Amount            float64   `json:"amount"`              // Số tiền thanh toán
	PaymentStatus     string    `json:"payment_status"`      // Trạng thái thanh toán (paid/unpaid)
	Email             string    `json:"email"`               // Email người dùng
	FullName          string    `json:"full_name"`           // Tên người dùng
	CancellationBy    string    `json:"cancellation_by"`     // "patient" hoặc "doctor"
	CancellationNote  string    `json:"cancellation_note"`   // Ghi chú lý do huỷ
	RefundAmount      float64   `json:"refund_amount"`       // Số tiền được hoàn
	RefundProcessDays int       `json:"refund_process_days"` // Số ngày xử lý hoàn tiền
	CancelledAt       time.Time `json:"cancelled_at"`        // Thời gian huỷ
}

// Hàm tạo BookingConfirmEvent
func CreateBookingConfirmEvent(
	userID, bookingID, expertID, email, fullName, doctorName string, doctorSpecialty []string,
	consultationDate, consultationTime string,
	duration int,
	consultationType, location, meetingLink string,
	amount float64,
	paymentStatus, bookingNotes, cancellationPolicy string,
) BookingConfirmEvent {
	return BookingConfirmEvent{
		UserID:             userID,
		BookingID:          bookingID,
		ExpertID:           expertID,
		Email:              email,
		FullName:           fullName,
		DoctorName:         doctorName,
		DoctorSpecialty:    doctorSpecialty,
		ConsultationDate:   consultationDate,
		ConsultationTime:   consultationTime,
		Duration:           duration,
		ConsultationType:   consultationType,
		Location:           location,
		MeetingLink:        meetingLink,
		Amount:             amount,
		PaymentStatus:      paymentStatus,
		BookingNotes:       bookingNotes,
		CancellationPolicy: cancellationPolicy,
		ConfirmedAt:        time.Now(),
	}
}
