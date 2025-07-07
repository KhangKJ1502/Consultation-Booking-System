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
	case "booking approve":
		return h.handleBookingApproveNotification(event)
	case "booking_confirmation":
		return h.handleBookingConfirmationNotification(event)
	case "booking_cancelled":
		return h.handleBookingCancelledNotification(event) // ✅ FIXED
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

func (h *EventHandler) handleBookingApproveNotification(event NotificationEvent) error {
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
	return h.emailService.SendConsultationBookingApprove(context.Background(), event.UserID, *bookingData)
}

func (h *EventHandler) handleBookingReminder(event NotificationEvent) error {
	log.Printf("⏰ Processing booking reminder for user: %s", event.UserID)
	// TODO: Implement booking reminder logic
	return nil
}

func (h *EventHandler) handleBookingCancelledNotification(event NotificationEvent) error {
	log.Printf("📧 Sending booking cancelled email to %s: %s", event.RecipientType, event.RecipientID)

	switch event.RecipientType {
	case "user":
		data := interfaces.ConsultationCancellationDataForUser{
			BookingID:         getString(event.Data["booking_id"]),
			DoctorName:        getString(event.Data["doctor_name"]),
			ConsultationDate:  getString(event.Data["consultation_date"]),
			ConsultationTime:  getString(event.Data["consultation_time"]),
			CancellationBy:    getString(event.Data["cancellation_by"]),
			CancellationNote:  getString(event.Data["cancellation_note"]),
			RefundAmount:      getFloat64(event.Data["refund_amount"]),
			RefundProcessDays: int(getFloat64(event.Data["refund_process_days"])),
		}

		if err := h.emailService.SendConsultationBookingCancelledForUser(context.Background(), event.RecipientID, data); err != nil {
			log.Printf("❌ Failed to send cancel email to user %s: %v", event.RecipientID, err)
		}

	case "expert":
		data := interfaces.ConsultationCancellationDataForExpert{
			BookingID:         getString(event.Data["booking_id"]),
			UserName:          getString(event.Data["user_name"]), // cần thêm vào khi publish
			ConsultationDate:  getString(event.Data["consultation_date"]),
			ConsultationTime:  getString(event.Data["consultation_time"]),
			CancellationBy:    getString(event.Data["cancellation_by"]),
			CancellationNote:  getString(event.Data["cancellation_note"]),
			RefundAmount:      getFloat64(event.Data["refund_amount"]),
			RefundProcessDays: int(getFloat64(event.Data["refund_process_days"])),
		}

		if err := h.emailService.SendConsultationBookingCancelledForExpert(context.Background(), event.RecipientID, data); err != nil {
			log.Printf("❌ Failed to send cancel email to expert %s: %v", event.RecipientID, err)
		}

	default:
		log.Printf("⚠️ Unknown recipient type: %s", event.RecipientType)
	}

	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getFloat64(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

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
