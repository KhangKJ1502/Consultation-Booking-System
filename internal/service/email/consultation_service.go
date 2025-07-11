package email

import (
	"context"
	"errors"
	"fmt"

	"cbs_backend/global"
	"cbs_backend/internal/service/interfaces"

	"go.uber.org/zap"
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

func (ces *ConsultationEmailService) SendBookingApprove(ctx context.Context, userID string, data interfaces.ConsultationBookingData) error {
	email := ces.userResolver.GetUserEmail(userID)
	if email == "" {
		global.Log.Error("User email not found", zap.String("userID", userID))
		return fmt.Errorf("user email not found")
	}

	template, err := ces.templateManager.GetTemplate("booking_created")
	if err != nil {
		global.Log.Error("Failed to get template", zap.Error(err))
		return ces.sendBookingConfirmationFallback(email, data)
	}
	//teamplateData bên phải là các thuộc tính trong db
	templateData := map[string]interface{}{
		"BookingID":          data.BookingID,
		"expert_name":        data.DoctorName,
		"DoctorSpecialty":    data.DoctorSpecialty,
		"booking_datetime":   data.ConsultationDate,
		"booking_time":       data.ConsultationTime,
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

	// Log template data để debug
	global.Log.Info("Rendering booking_created template", zap.Any("data", templateData))

	subject, body, err := ces.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		global.Log.Error("Get template failed render", zap.Error(err))
		return ces.sendBookingConfirmationFallback(email, data)
	}

	return ces.sender.Send(email, subject, body)
}

func (ces *ConsultationEmailService) SendBookingConfirmation(ctx context.Context, userID string, data interfaces.ConsultationBookingData) error {
	email := ces.userResolver.GetUserEmail(userID)
	template, err := ces.templateManager.GetTemplate("booking_confirmed")
	if err != nil {
		return ces.sendBookingConfirmationFallback(email, data)
	}

	global.Log.Error("flaied to get user email", zap.Error(err), zap.String("userID", userID))
	fmt.Printf("%s, %s, %d", data.BookingID, data.DoctorName, data.Duration)

	templateData := map[string]interface{}{
		"BookingID":          data.BookingID,
		"expert_name":        data.DoctorName,
		"DoctorSpecialty":    data.DoctorSpecialty,
		"booking_datetime":   data.ConsultationDate,
		"booking_time":       data.ConsultationTime,
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
func (ces *ConsultationEmailService) SendBookingCancelledForUser(ctx context.Context, userID string, data interfaces.ConsultationCancellationDataForUser) error {
	email := ces.userResolver.GetUserEmail(userID)
	if email == "" {
		global.Log.Error("failed to resolve user email", zap.String("userID", userID))
		return errors.New("user email not found")
	}

	template, err := ces.templateManager.GetTemplate("booking_cancelled")
	if err != nil {
		// return ces.sendBookingConfirmationFallback(email, data)
	}

	templateData := map[string]interface{}{
		"BookingID":         data.BookingID,
		"expert_name":       data.DoctorName,
		"booking_datetime":  data.ConsultationDate,
		"booking_time":      data.ConsultationTime,
		"CancellationNote":  data.CancellationNote,
		"RefundAmount":      FormatAmount(data.RefundAmount),
		"RefundProcessDays": data.RefundProcessDays,
		"CancellationBy":    data.CancellationBy,
	}

	subject, body, err := ces.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		// return ces.sendBookingConfirmationFallback(email, data)
	}

	return ces.sender.Send(email, subject, body)
}

func (ces *ConsultationEmailService) SendBookingCancelledForExpert(ctx context.Context, expertID string, data interfaces.ConsultationCancellationDataForExpert) error {
	email := ces.userResolver.GetDoctorEmail(expertID)
	if email == "" {
		global.Log.Error("failed to resolve expert email", zap.String("expertID", expertID))
		return errors.New("expert email not found")
	}

	template, err := ces.templateManager.GetTemplate("booking_cancelled_expert")
	if err != nil {
		// return ces.sendBookingConfirmationFallback(email, data)
	}
	templateData := map[string]interface{}{
		"BookingID":         data.BookingID,
		"expert_name":       data.UserName,
		"booking_datetime":  data.ConsultationDate,
		"booking_time":      data.ConsultationTime,
		"CancellationNote":  data.CancellationNote,
		"RefundAmount":      FormatAmount(data.RefundAmount),
		"RefundProcessDays": data.RefundProcessDays,
		"CancellationBy":    data.CancellationBy,
	}

	subject, body, err := ces.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		// return ces.sendBookingConfirmationFallback(email, data)
	}
	return ces.sender.Send(email, subject, body)
}

func (ces *ConsultationEmailService) SendReminderToUser(ctx context.Context, userID string, data interfaces.ConsultationReminderData) error {
	email := ces.userResolver.GetUserEmail(userID)
	template, err := ces.templateManager.GetTemplate("booking_reminder")
	if err != nil {
		// fallback gửi email text đơn giản
	}
	templateData := map[string]interface{}{
		"expert_name":      data.UserName,
		"BookingID":        data.BookingID,
		"ConsultationDate": data.ConsultationDate,
		"ConsultationTime": data.ConsultationTime,
		"MeetingLink":      data.MeetingLink,
		"Location":         data.Location,
		"TimeUntil":        data.TimeUntil,
	}
	subject, body, err := ces.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		// fallback
	}
	return ces.sender.Send(email, subject, body)
}

func (ces *ConsultationEmailService) sendReminderToExpert(ctx context.Context, userID string, data interfaces.ConsultationReminderData) error {
	email := ces.userResolver.GetUserEmail(userID)
	template, err := ces.templateManager.GetTemplate("booking_reminder_expert")
	if err != nil {
		// fallback gửi email text đơn giản
	}
	templateData := map[string]interface{}{
		"user_name":        data.UserName,
		"BookingID":        data.BookingID,
		"booking_datetime": data.ConsultationDate,
		"booking_time":     data.ConsultationTime,
		"MeetingLink":      data.MeetingLink,
		"Location":         data.Location,
		"TimeUntil":        data.TimeUntil,
	}
	subject, body, err := ces.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		// fallback
	}
	return ces.sender.Send(email, subject, body)
}

// ... implement other consultation methods

func (ces *ConsultationEmailService) sendBookingConfirmationFallback(email string, data interfaces.ConsultationBookingData) error {
	subject := "✅ Your 123 Consultation Booking is Confirmed"

	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #eee; border-radius: 8px;">
			<h2 style="color: #2c3e50;">📅 Booking Confirmation</h2>
			<p style="font-size: 16px;">Hello,</p>
			<p style="font-size: 16px;">Your consultation has been <strong>successfully booked</strong>! Below are the details:</p>

			<table style="width: 100%%; font-size: 16px; margin-top: 20px;">
				<tr>
					<td style="padding: 8px;"><strong>🔖 Booking ID:</strong></td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px;"><strong>👨‍⚕️ Doctor:</strong></td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px;"><strong>📆 Date:</strong></td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px;"><strong>⏰ Time:</strong></td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px;"><strong>💬 Type:</strong></td>
					<td style="padding: 8px;">%s</td>
				</tr>
			</table>

			<p style="margin-top: 30px; font-size: 15px; color: #555;">
				If you have any questions, feel free to contact us.  
				<br><br>
				Thank you for choosing our service! 🧑‍⚕️💙
			</p>
		</div>
	`, data.BookingID, data.DoctorName, data.ConsultationDate, data.ConsultationTime, data.ConsultationType)

	return ces.sender.Send(email, subject, body)
}
