package email

import (
	"context"
	"fmt"
)

type AuthEmailService struct {
	sender          *EmailSender
	templateManager *TemplateManager
	userResolver    *UserResolver
	baseURL         string
}

func NewAuthEmailService(
	sender *EmailSender,
	templateManager *TemplateManager,
	userResolver *UserResolver,
	baseURL string,
) *AuthEmailService {
	return &AuthEmailService{
		sender:          sender,
		templateManager: templateManager,
		userResolver:    userResolver,
		baseURL:         baseURL,
	}
}

func (aes *AuthEmailService) SendWelcomeEmail(ctx context.Context, userID string, email string, fullName string) error {
	template, err := aes.templateManager.GetTemplate("welcome_email")
	if err != nil {
		return aes.sendWelcomeEmailFallback(email, fullName, userID)
	}

	templateData := map[string]interface{}{
		"UserName":     fullName,
		"Email":        email,
		"UserID":       userID,
		"Platform":     "Consultation Booking System",
		"SupportEmail": "support@consultationbooking.com",
		"LoginURL":     fmt.Sprintf("%s/login", aes.baseURL),
		"ProfileURL":   fmt.Sprintf("%s/profile", aes.baseURL),
	}

	subject, body, err := aes.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		return aes.sendWelcomeEmailFallback(email, fullName, userID)
	}

	return aes.sender.Send(email, subject, body)
}

func (aes *AuthEmailService) SendEmailVerification(ctx context.Context, userID string, email string, verificationToken string) error {
	template, err := aes.templateManager.GetTemplate("email_verification")
	if err != nil {
		return aes.sendEmailVerificationFallback(email, verificationToken)
	}

	templateData := map[string]interface{}{
		"Email":           email,
		"VerificationURL": fmt.Sprintf("%s/verify-email?token=%s", aes.baseURL, verificationToken),
		"ExpiryHours":     "24",
	}

	subject, body, err := aes.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		return aes.sendEmailVerificationFallback(email, verificationToken)
	}

	return aes.sender.Send(email, subject, body)
}

func (aes *AuthEmailService) SendPasswordReset(ctx context.Context, email string, resetToken string) error {
	template, err := aes.templateManager.GetTemplate("password_reset")
	if err != nil {
		return aes.sendPasswordResetFallback(email, resetToken)
	}

	templateData := map[string]interface{}{
		"Email":       email,
		"ResetURL":    fmt.Sprintf("%s/reset-password?token=%s", aes.baseURL, resetToken),
		"ExpiryHours": "1",
	}

	subject, body, err := aes.templateManager.RenderTemplate(template, templateData)
	if err != nil {
		return aes.sendPasswordResetFallback(email, resetToken)
	}

	return aes.sender.Send(email, subject, body)
}

// Fallback methods
func (aes *AuthEmailService) sendWelcomeEmailFallback(email, fullName, userID string) error {
	subject := "Welcome to Consultation Booking System!"
	body := fmt.Sprintf(`
		<h2>Welcome %s!</h2>
		<p>Thank you for joining our Consultation Booking System.</p>
		<p>You can now book consultations with our qualified doctors.</p>
		<p><a href="%s/login">Login to your account</a></p>
		<p>If you have any questions, please contact our support team.</p>
	`, fullName, aes.baseURL)

	return aes.sender.Send(email, subject, body)
}

func (aes *AuthEmailService) sendEmailVerificationFallback(email, verificationToken string) error {
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<h2>Email Verification Required</h2>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="%s/verify-email?token=%s">Verify Email</a></p>
		<p>This link will expire in 24 hours.</p>
	`, aes.baseURL, verificationToken)

	return aes.sender.Send(email, subject, body)
}

func (aes *AuthEmailService) sendPasswordResetFallback(email, resetToken string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<h2>Password Reset</h2>
		<p>Click the link below to reset your password:</p>
		<p><a href="%s/reset-password?token=%s">Reset Password</a></p>
		<p>This link will expire in 1 hour.</p>
	`, aes.baseURL, resetToken)

	return aes.sender.Send(email, subject, body)
}
