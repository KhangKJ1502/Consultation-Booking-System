package kafka

import (
	"encoding/json"
	"fmt"
	"time"
)

// =============================================================================
// USER EVENT PUBLISHERS
// =============================================================================

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

// =============================================================================
// NOTIFICATION EVENT PUBLISHERS
// =============================================================================

func PublishNotificationEvent(event NotificationEvent) error {
	event.CreatedAt = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification event: %v", err)
	}

	return Publish("user-notifications", data)
}

// =============================================================================
// BOOKING EVENT PUBLISHERS
// =============================================================================

func PublishBookingEvent(event BookingEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal booking event: %w", err)
	}

	return Publish("booking-events", eventData)
}

func PublishBookingCreatedEvent(event BookingCreatedEvent) error {
	event.EventType = "booking_created"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal booking created event: %v", err)
	}

	return Publish("booking-events", data)
}

func PublishBookingConfirmEvent(event BookingConfirmEvent) error {
	event.EventType = "booking_confirm"
	event.ConfirmedAt = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal booking confirm event: %v", err)
	}

	return Publish("booking-events", data)
}

func PublishBookingUpdatedEvent(event BookingEvent) error {
	event.EventType = "booking_updated"
	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal booking updated event: %v", err)
	}

	return Publish("booking-events", data)
}

func PublishBookingCancelledEvent(event BookingEvent) error {
	event.EventType = "booking_cancelled"
	event.Timestamp = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal booking cancelled event: %v", err)
	}

	return Publish("booking-events", data)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func CreateBookingConfirmEvent(
	userID, bookingID, expertID, email, fullName, doctorName string,
	consultationDate, consultationTime string,
	duration int,
	consultationType, location, meetingLink string,
	amount float64,
	paymentStatus, bookingNotes, cancellationPolicy string,
) BookingConfirmEvent {
	return BookingConfirmEvent{
		EventType:          "booking_confirm",
		UserID:             userID,
		BookingID:          bookingID,
		ExpertID:           expertID,
		Email:              email,
		FullName:           fullName,
		DoctorName:         doctorName,
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
