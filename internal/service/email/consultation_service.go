package email

import (
	"context"
	"fmt"

	"cbs_backend/internal/service/interfaces"
)

type ConsultationEmailService struct {
	sender          *EmailSender
	templateManager *TemplateManager
	userResolver    *UserResolver
	baseURL         string
}

func NewConsultationEmailService(
	sender *EmailSender,
	templateManager *TemplateManager,
	userResolver *UserResolver,
	baseURL string,
) *ConsultationEmailService {
	return &ConsultationEmailService{
		sender:          sender,
		templateManager: templateManager,
		userResolver:    userResolver,
		baseURL:         baseURL,
	}
}

func (ces *ConsultationEmailService) SendBookingConfirmation(ctx context.Context, userID string, data interfaces.ConsultationBookingData) error {
	email := ces.userResolver.GetUserEmail(userID)
	template, err := ces.templateManager.GetTemplate("consultation_booking_confirmation")
	if err != nil {
		return ces.sendBookingConfirmationFallback(email, data)
	}

	templateData := map[string]interface{}{
		"BookingID":          data.BookingID,
		"DoctorName":         data.DoctorName,
		"DoctorSpecialty":    data.DoctorSpecialty,
		"ConsultationDate":   data.ConsultationDate,
		"ConsultationTime":   data.ConsultationTime,
		"Duration":           data.Duration,
		"ConsultationType":   data.ConsultationType,
		"Location":           data.Location,
		"MeetingLink":        data.MeetingLink,
		"Amount":             FormatAmount(data.Amount),
		"PaymentStatus":      data.PaymentStatus,
		"BookingNotes":       data.BookingNotes,
		"CancellationPolicy": data.CancellationPolicy,
		"BookingURL":         fmt.Sprintf("%s/bookings/%s", ces.baseURL, data.BookingID),
	}

	subject, body, err := ces.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		return ces.sendBookingConfirmationFallback(email, data)
	}

	return ces.sender.Send(email, subject, body)
}

// ... implement other consultation methods

func (ces *ConsultationEmailService) sendBookingConfirmationFallback(email string, data interfaces.ConsultationBookingData) error {
	subject := "Consultation Booking Confirmed"
	body := fmt.Sprintf(`
		<h2>Booking Confirmation</h2>
		<p>Your consultation has been booked successfully!</p>
		<p><strong>Booking ID:</strong> %s</p>
		<p><strong>Doctor:</strong> %s</p>
		<p><strong>Date:</strong> %s at %s</p>
		<p><strong>Type:</strong> %s</p>
	`, data.BookingID, data.DoctorName, data.ConsultationDate, data.ConsultationTime, data.ConsultationType)

	return ces.sender.Send(email, subject, body)
}
