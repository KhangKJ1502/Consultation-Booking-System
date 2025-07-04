package kafka

import (
	"cbs_backend/internal/service/interfaces"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// =============================================================================
// NOTIFICATION EVENT HANDLERS
// =============================================================================

func (h *EventHandler) handleNotificationEvent(data []byte) error {
	var event NotificationEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal notification event: %v", err)
		return err
	}

	log.Printf("📧 Processing notification for user %s: %s", event.UserID, event.Type)

	switch event.Type {
	case "welcome_email":
		return h.handleWelcomeEmail(event)
	case "profile_updated":
		return h.handleProfileUpdatedNotification(event)
	case "booking_confirmation":
		return h.handleBookingConfirmationNotification(event)
	default:
		log.Printf("⚠️ Unknown notification type: %s", event.Type)
	}

	return nil
}

// =============================================================================
// SPECIFIC NOTIFICATION HANDLERS
// =============================================================================

func (h *EventHandler) handleWelcomeEmail(event NotificationEvent) error {
	log.Printf("📬 Sending welcome email to user: %s", event.UserID)

	// Extract user data from event
	email, _ := event.Data["email"].(string)
	fullName, _ := event.Data["full_name"].(string)
	userID, _ := event.Data["user_id"].(string)

	// Check if email service is available
	if h.emailService == nil {
		log.Printf("⚠️ EmailService is nil - running in simulation mode")
		log.Printf("✅ [SIMULATION] Welcome email sent to %s (%s)", email, fullName)
		return nil
	}

	// Send real email
	log.Printf("✅ EmailService available - sending real email")
	return h.emailService.SendWelcomeEmail(context.Background(), userID, email, fullName)
}

func (h *EventHandler) handleProfileUpdatedNotification(event NotificationEvent) error {
	log.Printf("📬 Processing profile updated notification for user: %s", event.UserID)

	// Could send email, push notification, etc.
	log.Printf("✅ Profile updated notification processed for user: %s", event.UserID)
	return nil
}

func (h *EventHandler) handleBookingConfirmationNotification(event NotificationEvent) error {
	log.Printf("📬 Sending booking confirmation email to user: %s", event.UserID)

	// Extract booking data from event
	bookingData := h.extractBookingDataFromEvent(event)
	if bookingData == nil {
		log.Printf("❌ Failed to extract booking data from event")
		return fmt.Errorf("failed to extract booking data")
	}

	// Check if email service is available
	if h.emailService == nil {
		log.Printf("⚠️ EmailService is nil - running in simulation mode")
		log.Printf("✅ [SIMULATION] Booking confirmation email sent to user %s for booking %s",
			event.UserID, bookingData.BookingID)
		return nil
	}

	// Send real email
	log.Printf("✅ EmailService available - sending real booking confirmation email")
	return h.emailService.SendConsultationBookingConfirmation(context.Background(), event.UserID, *bookingData)
}

func (h *EventHandler) handleBookingReminder(event NotificationEvent) error {
	log.Printf("⏰ Processing booking reminder for user: %s", event.UserID)
	// TODO: Implement booking reminder logic
	return nil
}

func (h *EventHandler) handleBookingCancelledNotification(event NotificationEvent) error {
	log.Printf("❌ Processing booking cancelled notification for user: %s", event.UserID)
	// TODO: Implement booking cancelled notification logic
	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func (h *EventHandler) extractBookingDataFromEvent(event NotificationEvent) *interfaces.ConsultationBookingData {
	data := event.Data
	if data == nil {
		return nil
	}

	// Helper functions for safe type conversion
	getString := func(key string) string {
		if val, ok := data[key].(string); ok {
			return val
		}
		return ""
	}

	getFloat64 := func(key string) float64 {
		if val, ok := data[key].(float64); ok {
			return val
		}
		return 0.0
	}

	getInt := func(key string) int {
		if val, ok := data[key].(int); ok {
			return val
		}
		if val, ok := data[key].(float64); ok {
			return int(val)
		}
		return 0
	}

	// Extract booking data
	bookingData := &interfaces.ConsultationBookingData{
		BookingID:          getString("booking_id"),
		DoctorName:         getString("doctor_name"),
		DoctorSpecialty:    getString("doctor_specialty"),
		ConsultationDate:   getString("consultation_date"),
		ConsultationTime:   getString("consultation_time"),
		Duration:           getInt("duration"),
		ConsultationType:   getString("consultation_type"),
		Location:           getString("location"),
		MeetingLink:        getString("meeting_link"),
		Amount:             getFloat64("amount"),
		PaymentStatus:      getString("payment_status"),
		BookingNotes:       getString("booking_notes"),
		CancellationPolicy: getString("cancellation_policy"),
	}

	// Validate required fields
	if bookingData.BookingID == "" || bookingData.DoctorName == "" {
		log.Printf("❌ Missing required booking data fields")
		return nil
	}

	return bookingData
}
