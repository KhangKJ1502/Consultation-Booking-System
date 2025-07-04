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
	UserID    string                 `json:"user_id"`
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
	EventType          string  `json:"event_type"`
	UserID             string  `json:"user_id"`
	BookingID          string  `json:"booking_id"`
	DoctorName         string  `json:"doctor_name"`
	DoctorSpecialty    string  `json:"doctor_specialty"`
	ConsultationDate   string  `json:"consultation_date"`
	ConsultationTime   string  `json:"consultation_time"`
	Duration           string  `json:"duration"`
	ConsultationType   string  `json:"consultation_type"`
	Location           string  `json:"location"`
	MeetingLink        string  `json:"meeting_link"`
	Amount             float64 `json:"amount"`
	PaymentStatus      string  `json:"payment_status"`
	BookingNotes       string  `json:"booking_notes"`
	CancellationPolicy string  `json:"cancellation_policy"`
	Email              string  `json:"email"`
	FullName           string  `json:"full_name"`
}

// BookingConfirmEvent
type BookingConfirmEvent struct {
	EventType          string    `json:"event_type"`
	UserID             string    `json:"user_id"`
	BookingID          string    `json:"booking_id"`
	ExpertID           string    `json:"expert_id"`
	DoctorName         string    `json:"doctor_name"`
	DoctorSpecialty    string    `json:"doctor_specialty"`
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
