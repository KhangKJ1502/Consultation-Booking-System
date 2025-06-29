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
