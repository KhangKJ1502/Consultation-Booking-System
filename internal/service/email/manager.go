package email

import (
	"context"

	"cbs_backend/internal/service/interfaces"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EmailManager is the main service that implements interfaces.EmailService
type EmailManager struct {
	authService         *AuthEmailService
	consultationService *ConsultationEmailService
	// paymentService      *PaymentEmailService
	// doctorService       *DoctorEmailService
	// systemService       *SystemEmailService
}

func NewEmailManager(db *gorm.DB, logger *zap.Logger) interfaces.EmailService {
	config := LoadEmailConfig()

	// Initialize core components
	sender := NewEmailSender(config, logger)
	templateManager := NewTemplateManager(db, logger)
	userResolver := NewUserResolver(db, logger)

	// Initialize domain services
	authService := NewAuthEmailService(sender, templateManager, userResolver, config.BaseURL)
	consultationService := NewConsultationEmailService(sender, templateManager, userResolver, config.BaseURL)

	return &EmailManager{
		authService:         authService,
		consultationService: consultationService,
	}
}

// Implement interfaces.EmailService by delegating to appropriate services
func (em *EmailManager) SendWelcomeEmail(ctx context.Context, userID string, email string, fullName string) error {
	return em.authService.SendWelcomeEmail(ctx, userID, email, fullName)
}

func (em *EmailManager) SendEmailVerification(ctx context.Context, userID string, email string, verificationToken string) error {
	return em.authService.SendEmailVerification(ctx, userID, email, verificationToken)
}

func (em *EmailManager) SendPasswordReset(ctx context.Context, email string, resetToken string) error {
	return em.authService.SendPasswordReset(ctx, email, resetToken)
}

func (em *EmailManager) SendConsultationBookingConfirmation(ctx context.Context, userID string, data interfaces.ConsultationBookingData) error {
	return em.consultationService.SendBookingConfirmation(ctx, userID, data)
}
func (em *EmailManager) SendConsultationBookingApprove(ctx context.Context, userID string, data interfaces.ConsultationBookingData) error {
	return em.consultationService.SendBookingApprove(ctx, userID, data)
}
func (em *EmailManager) SendConsultationBookingCancelledForUser(ctx context.Context, userID string, data interfaces.ConsultationCancellationDataForUser) error {
	return em.consultationService.SendBookingCancelledForUser(ctx, userID, data)
}
func (em *EmailManager) SendConsultationBookingCancelledForExpert(ctx context.Context, userID string, data interfaces.ConsultationCancellationDataForExpert) error {
	return em.consultationService.SendBookingCancelledForExpert(ctx, userID, data)
}

// ... implement other interface methods
