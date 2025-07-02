package kafka

import (
	"encoding/json"
	"fmt"
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

// Event Publishers
func PublishUserRegisteredEvent(event UserRegisteredEvent) error {
	event.EventType = "user_registered"
	event.RegisteredAt = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user registered event: %v", err)
	}

	return Publish("user-events", data)
}

func PublishUserProfileUpdatedEvent(event UserProfileUpdatedEvent) error {
	event.EventType = "user_profile_updated"
	event.UpdatedAt = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile updated event: %v", err)
	}

	return Publish("user-events", data)
}

func PublishNotificationEvent(event NotificationEvent) error {
	event.CreatedAt = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification event: %v", err)
	}

	return Publish("user-notifications", data)
}

func PublishBookingEvent(event BookingEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal booking event: %w", err)
	}
	topic := "booking-events"
	return Publish(topic, eventData)
}
