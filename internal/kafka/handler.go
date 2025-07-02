package kafka

import (
	"cbs_backend/internal/service/interfaces"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type EventHandler struct {
	emailService interfaces.EmailService // ← Đã có field này
}

// Constructor cũ - chỉ để backward compatibility
func NewEventHandler() *EventHandler {
	return &EventHandler{
		emailService: nil, // Will run in simulation mode
	}
}

// Constructor mới - với EmailService
func NewEventHandlerWithEmailService(emailService interfaces.EmailService) *EventHandler {
	return &EventHandler{
		emailService: emailService, // ← Truyền emailService vào
	}
}

// HandleMessage - Entry point cho tất cả messages
func (h *EventHandler) HandleMessage(message *sarama.ConsumerMessage) error {
	log.Printf("📨 Received message from topic: %s, partition: %d, offset: %d",
		message.Topic, message.Partition, message.Offset)

	// Topic routing
	switch message.Topic {
	case "user-events":
		return h.handleUserEvent(message.Value)
	case "user-notifications":
		return h.handleNotificationEvent(message.Value)
	case "booking-events":
		return h.handleBoongkingsEvent(message.Value)
	default:
		log.Printf("⚠️ Unknown topic: %s", message.Topic)
	}

	return nil
}

// handleUserEvent - Xử lý user events
func (h *EventHandler) handleUserEvent(data []byte) error {
	log.Printf("👤 Raw user event: %s", string(data))

	var userEvent map[string]interface{}
	if err := json.Unmarshal(data, &userEvent); err != nil {
		log.Printf("❌ Failed to unmarshal user event: %v", err)
		return err
	}

	// Detect event type
	eventType, exists := userEvent["event_type"].(string)
	if !exists {
		log.Printf("⚠️ User event missing event_type field")
		return nil
	}

	switch eventType {
	case "user_registered":
		var event UserRegisteredEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return h.handleUserRegistered(event)
	case "user_profile_updated":
		var event UserProfileUpdatedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		return h.handleUserProfileUpdated(event)
	}

	return nil
}

// handleUserRegistered - Xử lý khi user đăng ký
func (h *EventHandler) handleUserRegistered(event UserRegisteredEvent) error {
	log.Printf("🆕 New user registered: %s (%s)", event.Email, event.UserID)

	// Tạo notification để gửi welcome email
	notification := NotificationEvent{
		UserID:  event.UserID,
		Type:    "welcome_email",
		Title:   "Welcome to Consultation Booking System",
		Message: fmt.Sprintf("Welcome %s! Your account has been created successfully.", event.FullName),
		Data: map[string]interface{}{
			"user_id":   event.UserID,
			"email":     event.Email,
			"full_name": event.FullName,
		},
	}

	// Gửi notification event
	return PublishNotificationEvent(notification)
}

// handleUserProfileUpdated - Xử lý khi user cập nhật profile
func (h *EventHandler) handleUserProfileUpdated(event UserProfileUpdatedEvent) error {
	log.Printf("🔄 User profile updated: %s", event.UserID)

	// Có thể thêm logic gửi email thông báo profile updated
	notification := NotificationEvent{
		UserID:  event.UserID,
		Type:    "profile_updated",
		Title:   "Profile Updated",
		Message: "Your profile has been updated successfully.",
		Data: map[string]interface{}{
			"user_id": event.UserID,
			"changes": event.Changes,
		},
	}

	return PublishNotificationEvent(notification)
}

// handleNotificationEvent - Xử lý notification events
func (h *EventHandler) handleNotificationEvent(data []byte) error {
	var event NotificationEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal notification event: %v", err)
		return err
	}

	log.Printf("📧 Processing notification for user %s: %s", event.UserID, event.Type)

	// Xử lý theo type notification
	switch event.Type {
	case "welcome_email":
		return h.handleWelcomeEmail(event)
	case "profile_updated":
		return h.handleProfileUpdatedNotification(event)
	default:
		log.Printf("⚠️ Unknown notification type: %s", event.Type)
	}

	return nil
}

// handleWelcomeEmail - Xử lý gửi welcome email
func (h *EventHandler) handleWelcomeEmail(event NotificationEvent) error {
	log.Printf("📬 Sending welcome email to user: %s", event.UserID)

	// Extract data from event
	email, _ := event.Data["email"].(string)
	fullName, _ := event.Data["full_name"].(string)
	userID, _ := event.Data["user_id"].(string)

	// Debug log để kiểm tra emailService
	if h.emailService == nil {
		log.Printf("⚠️ EmailService is nil - running in simulation mode")
		// Simulation - Log thay vì gửi email thật
		log.Printf("✅ [SIMULATION] Welcome email sent to %s (%s)", email, fullName)
		return nil
	}

	// Gửi email thật thông qua emailService
	log.Printf("✅ EmailService available - sending real email")
	return h.emailService.SendWelcomeEmail(context.Background(), userID, email, fullName)
}

// handleProfileUpdatedNotification - Xử lý thông báo profile updated
func (h *EventHandler) handleProfileUpdatedNotification(event NotificationEvent) error {
	log.Printf("📬 Processing profile updated notification for user: %s", event.UserID)

	// Có thể gửi email, push notification, etc.
	log.Printf("✅ Profile updated notification processed for user: %s", event.UserID)
	return nil
}

// /---------------------------- Booking Service ---------------------\
// handleBookingEvent - Xử lý Booking events
func (h *EventHandler) handleBoongkingsEvent(data []byte) error {
	log.Printf("📝 Raw booking event: %s", string(data))

	// First try to parse as direct booking event
	var bookingEvent map[string]interface{}
	if err := json.Unmarshal(data, &bookingEvent); err != nil {
		log.Printf("❌ Failed to unmarshal booking event: %v", err)
		return err
	}

	// Check if it's a direct booking event or notification event
	if eventType, exists := bookingEvent["event_type"].(string); exists {
		// Direct booking event format
		switch eventType {
		case "booking_created":
			return h.handleBookingCreated(data)
		// case "booking_updated":
		// 	return h.handleBookingUpdated(data)
		// case "booking_cancelled":
		// 	return h.handleBookingCancelled(data)
		default:
			log.Printf("⚠️ Unknown booking event type: %s", eventType)
		}
	} else {
		// Try notification event format
		var event NotificationEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("❌ Failed to unmarshal as notification event: %v", err)
			return err
		}

		log.Printf("📧 Processing booking notification for user %s: %s", event.UserID, event.Type)

		// Xử lý theo type notification
		// switch event.Type {
		// case "booking_confirmation":
		// 	return h.handleBookingConfirmation(event)
		// case "booking_reminder":
		// 	return h.handleBookingReminder(event)
		// case "booking_cancelled":
		// 	return h.handleBookingCancelledNotification(event)
		// default:
		// 	log.Printf("⚠️ Unknown booking notification type: %s", event.Type)
		// }
	}

	return nil
}

// handleBookingCreated - Xử lý khi booking được tạo
func (h *EventHandler) handleBookingCreated(data []byte) error {
	var event BookingCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal booking created event: %v", err)
		return err
	}

	log.Printf("📝 New booking created: %s for user %s", event.BookingID, event.UserID)

	// Tạo notification để gửi confirmation email
	notification := NotificationEvent{
		UserID:  event.UserID,
		Type:    "booking_confirmation",
		Title:   "Booking Confirmation",
		Message: fmt.Sprintf("Your consultation booking %s has been confirmed.", event.BookingID),
		Data: map[string]interface{}{
			"user_id":             event.UserID,
			"booking_id":          event.BookingID,
			"doctor_name":         event.DoctorName,
			"doctor_specialty":    event.DoctorSpecialty,
			"consultation_date":   event.ConsultationDate,
			"consultation_time":   event.ConsultationTime,
			"duration":            event.Duration,
			"consultation_type":   event.ConsultationType,
			"location":            event.Location,
			"meeting_link":        event.MeetingLink,
			"amount":              event.Amount,
			"payment_status":      event.PaymentStatus,
			"booking_notes":       event.BookingNotes,
			"cancellation_policy": event.CancellationPolicy,
			"email":               event.Email,
			"full_name":           event.FullName,
		},
	}

	// Publish notification event để gửi email
	return PublishNotificationEvent(notification)
}
