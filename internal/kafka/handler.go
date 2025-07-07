package kafka

import (
	"cbs_backend/internal/service/interfaces"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// =============================================================================
// EVENT HANDLER
// =============================================================================

type EventHandler struct {
	emailService interfaces.EmailService
}

// Constructors
func NewEventHandler() *EventHandler {
	return &EventHandler{
		emailService: nil, // Will run in simulation mode
	}
}

func NewEventHandlerWithEmailService(emailService interfaces.EmailService) *EventHandler {
	return &EventHandler{
		emailService: emailService,
	}
}

// =============================================================================
// MAIN MESSAGE HANDLER
// =============================================================================

func (h *EventHandler) HandleMessage(message *sarama.ConsumerMessage) error {
	log.Printf("📨 Received message from topic: %s, partition: %d, offset: %d",
		message.Topic, message.Partition, message.Offset)
	// Route to appropriate handler based on topic
	switch message.Topic {
	case "user-events":
		return h.handleUserEvent(message.Value)
	case "user-notifications":
		return h.handleNotificationEvent(message.Value)
	case "booking-events":
		return h.handleBookingEvent(message.Value)
	default:
		log.Printf("⚠️ Unknown topic: %s", message.Topic)
	}

	return nil
}

// =============================================================================
// USER EVENT HANDLERS
// =============================================================================

func (h *EventHandler) handleUserEvent(data []byte) error {
	log.Printf("👤 Raw user event: %s", string(data))

	var userEvent map[string]interface{}
	if err := json.Unmarshal(data, &userEvent); err != nil {
		log.Printf("❌ Failed to unmarshal user event: %v", err)
		return err
	}

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
	default:
		log.Printf("⚠️ Unknown user event type: %s", eventType)
	}

	return nil
}

func (h *EventHandler) handleUserRegistered(event UserRegisteredEvent) error {
	log.Printf("🆕 New user registered: %s (%s)", event.Email, event.UserID)

	// Create welcome email notification
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

	return PublishNotificationEvent(notification)
}

func (h *EventHandler) handleUserProfileUpdated(event UserProfileUpdatedEvent) error {
	log.Printf("🔄 User profile updated: %s", event.UserID)

	// Create profile updated notification
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

// =============================================================================
// BOOKING EVENT HANDLERS
// =============================================================================

func (h *EventHandler) handleBookingEvent(data []byte) error {
	log.Printf("📝 Raw booking event: %s", string(data))

	var bookingEvent map[string]interface{}
	if err := json.Unmarshal(data, &bookingEvent); err != nil {
		log.Printf("❌ Failed to unmarshal booking event: %v", err)
		return err
	}

	// Check if it's a direct booking event
	if eventType, exists := bookingEvent["event_type"].(string); exists {
		switch eventType {
		case "booking_confirmation":
			return h.handleBookingCreated(data)
		case "booking_confirm":
			return h.handleBookingConfirmed(data)
		case "booking_updated":
			return h.handleBookingUpdated(data)
		case "booking_cancelled":
			return h.handleBookingCancelled(data)
		default:
			log.Printf("⚠️ Unknown booking event type: %s", eventType)
		}
	} else {
		// Try as notification event
		var event NotificationEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("❌ Failed to unmarshal as notification event: %v", err)
			return err
		}

		return h.handleBookingNotification(event)
	}

	return nil
}

func (h *EventHandler) handleBookingCreated(data []byte) error {
	var event BookingCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal booking created event: %v", err)
		return err
	}

	log.Printf("📝 New booking created: %s for user %s", event.BookingID, event.UserID)

	// Create booking confirmation notification
	notification := NotificationEvent{
		UserID:  event.UserID,
		Type:    "booking_confirmation",
		Title:   "Booking Confirmation",
		Message: fmt.Sprintf("Your consultation booking %s has been confirmed.", event.BookingID),
		Data:    h.createBookingNotificationData(event),
	}

	return PublishNotificationEvent(notification)
}

func (h *EventHandler) handleBookingConfirmed(data []byte) error {
	var event BookingConfirmEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal booking confirm event: %v", err)
		return err
	}

	log.Printf("✅ Booking confirmed: %s for user %s by expert %s",
		event.BookingID, event.UserID, event.ExpertID)

	// Create expert confirmation notification
	notification := NotificationEvent{
		UserID: event.UserID,
		Type:   "booking approve",
		Title:  "Booking Confirmed by Expert",
		Message: fmt.Sprintf("Your consultation booking %s has been confirmed by %s.",
			event.BookingID, event.DoctorName),
		Data: h.createBookingConfirmNotificationData(event), // chỗ này là dùng dữ liệu sau khi đã map sau đó gửi vào struct NotificationEven
	}

	return PublishNotificationEvent(notification)
}

func (h *EventHandler) handleBookingUpdated(data []byte) error {
	log.Printf("🔄 Booking updated event received")
	// TODO: Implement booking updated logic
	return nil
}
func (h *EventHandler) handleBookingCancelled(data []byte) error {
	var event BookingCancelledEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal booking cancel event: %v", err)
		return err
	}

	log.Printf("📪 Booking cancelled: %s by %s (user: %s, expert: %s)",
		event.BookingID, event.CancellationBy, event.UserID, event.ExpertID)

	notificationData := h.createBookingCancelNotificationData(event)

	// Notification for User
	notificationForUser := NotificationEvent{
		UserID:        event.UserID,
		RecipientID:   event.UserID,
		RecipientType: "user",
		Type:          "booking_cancelled",
		Title:         "Lịch hẹn bị huỷ",
		Message:       fmt.Sprintf("Lịch hẹn với chuyên gia %s vào %s đã bị huỷ.", event.DoctorName, event.ConsultationDate),
		Data:          notificationData,
		CreatedAt:     time.Now(),
	}

	// Notification for Expert
	notificationForExpert := NotificationEvent{
		UserID:        event.UserID, // ai là người gây ra thì để đây (nếu chuyên gia huỷ thì gán expertID)
		RecipientID:   event.ExpertID,
		RecipientType: "expert",
		Type:          "booking_cancelled",
		Title:         "Lịch hẹn bị huỷ",
		Message:       fmt.Sprintf("Người dùng %s đã huỷ lịch hẹn vào %s.", event.FullName, event.ConsultationDate),
		Data:          notificationData,
		CreatedAt:     time.Now(),
	}

	// Gửi cả 2 notification qua Kafka
	if err := PublishNotificationEvent(notificationForUser); err != nil {
		log.Printf("❌ Failed to publish user notification: %v", err)
	}
	if err := PublishNotificationEvent(notificationForExpert); err != nil {
		log.Printf("❌ Failed to publish expert notification: %v", err)
	}

	return nil
}

func (h *EventHandler) handleBookingNotification(event NotificationEvent) error {
	log.Printf("📧 Processing booking notification for user %s: %s", event.UserID, event.Type)

	switch event.Type {
	case "booking_confirmed":
		return h.handleBookingConfirmationNotification(event)
	case "booking_reminder":
		return h.handleBookingReminder(event)
	case "booking_cancelled":
		return h.handleBookingCancelledNotification(event)
	default:
		log.Printf("⚠️ Unknown booking notification type: %s", event.Type)
	}

	return nil
}

// =============================================================================
// HELPER FUNCTIONS FOR BOOKING DATA
// =============================================================================
// Hàm này để Map đối tượng từ Booking sang kiểu của thông báo
func (h *EventHandler) createBookingNotificationData(event BookingCreatedEvent) map[string]interface{} {
	return map[string]interface{}{
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
	}
}

func (h *EventHandler) createBookingConfirmNotificationData(event BookingConfirmEvent) map[string]interface{} {
	return map[string]interface{}{
		"user_id":             event.UserID,
		"booking_id":          event.BookingID,
		"expert_id":           event.ExpertID,
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
		"confirmed_at":        event.ConfirmedAt,
	}
}

func (h *EventHandler) createBookingCancelNotificationData(event BookingCancelledEvent) map[string]interface{} {
	return map[string]interface{}{
		"booking_id":          event.BookingID,
		"consultation_date":   event.ConsultationDate,
		"consultation_time":   event.ConsultationTime,
		"doctor_name":         event.DoctorName,
		"doctor_specialty":    event.DoctorSpecialty,
		"consultation_type":   event.ConsultationType,
		"location":            event.Location,
		"meeting_link":        event.MeetingLink,
		"cancellation_by":     event.CancellationBy,
		"cancellation_note":   event.CancellationNote,
		"refund_amount":       event.RefundAmount,
		"refund_process_days": event.RefundProcessDays,
		"cancelled_at":        event.CancelledAt.Format(time.RFC3339),
	}
}
