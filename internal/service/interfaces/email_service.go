// internal/service/interfaces/email_service.go
package interfaces

import "context"

// EmailService interface định nghĩa các phương thức email cần thiết
type EmailService interface {
	// User-related emails
	SendWelcomeEmail(ctx context.Context, userID string, email string, fullName string) error
	SendEmailVerification(ctx context.Context, userID string, email string, verificationToken string) error
	SendPasswordReset(ctx context.Context, email string, resetToken string) error

	// // Consultation-related emails
	SendConsultationBookingConfirmation(ctx context.Context, userID string, data ConsultationBookingData) error
	// SendConsultationReminder(ctx context.Context, userID string, data ConsultationReminderData) error
	// SendConsultationCancellation(ctx context.Context, userID string, data ConsultationCancellationData) error
	// SendConsultationRescheduled(ctx context.Context, userID string, data ConsultationRescheduleData) error

	// // Doctor-related emails
	// SendNewConsultationToDoctor(ctx context.Context, doctorID string, data ConsultationNotificationData) error
	// SendConsultationCancellationToDoctor(ctx context.Context, doctorID string, data ConsultationCancellationData) error

	// // // Payment-related emails
	// SendPaymentConfirmation(ctx context.Context, userID string, data PaymentConfirmationData) error
	// SendPaymentFailed(ctx context.Context, userID string, data PaymentFailedData) error
	// SendRefundConfirmation(ctx context.Context, userID string, data RefundConfirmationData) error

	// // // System notifications

	// // SendSystemMaintenance(ctx context.Context, emails []string, data MaintenanceData) error
	// SendNewsletterEmail(ctx context.Context, emails []string, data NewsletterData) error
}
