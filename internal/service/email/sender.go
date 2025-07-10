package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
)

type EmailSender struct {
	config EmailConfig
	logger *zap.Logger
}

func NewEmailSender(config EmailConfig, logger *zap.Logger) *EmailSender {
	return &EmailSender{
		config: config,
		logger: logger,
	}
}

func (es *EmailSender) Send(to, subject, body string) error {
	if to == "" {
		es.logger.Error("Recipient email is empty")
		return fmt.Errorf("recipient email is required")
	}

	es.logger.Info("Sending email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("from", es.config.FromEmail),
		zap.String("smtp_host", es.config.SMTPHost),
	)

	auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)

	msg := []string{
		fmt.Sprintf("From: %s <%s>", es.config.FromName, es.config.FromEmail),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}

	message := []byte(strings.Join(msg, "\r\n"))
	addr := fmt.Sprintf("%s:%s", es.config.SMTPHost, es.config.SMTPPort)

	if err := smtp.SendMail(addr, auth, es.config.FromEmail, []string{to}, message); err != nil {
		es.logger.Error("Failed to send email", zap.Error(err), zap.String("to", to))
		return fmt.Errorf("failed to send email: %w", err)
	}

	es.logger.Info("Email sent successfully", zap.String("to", to))
	return nil
}

func (es *EmailSender) SendBulk(emails []string, subject, body string) error {
	for _, email := range emails {
		if err := es.Send(email, subject, body); err != nil {
			es.logger.Error("Failed to send bulk email", zap.Error(err), zap.String("email", email))
			continue
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

// SendEmail - Enhanced version với template support
func (es *EmailSender) SendEmail(recipient, subject, body, templateName string, data interface{}) error {
	// Validate input
	if err := es.validateInput(recipient, subject, body); err != nil {
		return err
	}

	es.logger.Info("Sending email",
		zap.String("to", recipient),
		zap.String("subject", subject),
		zap.String("from", es.config.FromEmail),
		zap.String("smtp_host", es.config.SMTPHost),
		zap.String("template", templateName),
	)

	// Build message
	message, err := es.buildMessage(recipient, subject, body)
	if err != nil {
		es.logger.Error("Failed to build email message", zap.Error(err))
		return fmt.Errorf("failed to build email message: %w", err)
	}

	// Send via SMTP
	if err := es.sendViaSMTP(recipient, message); err != nil {
		es.logger.Error("Failed to send email", zap.Error(err), zap.String("to", recipient))
		return fmt.Errorf("failed to send email: %w", err)
	}

	es.logger.Info("Email sent successfully", zap.String("to", recipient))
	return nil
}

// validateInput kiểm tra input data

func (es *EmailSender) validateInput(recipient, subject, body string) error {
	if recipient == "" {
		es.logger.Error("Recipient email is empty")
		return fmt.Errorf("recipient email is required")
	}

	if subject == "" {
		es.logger.Error("Subject is empty")
		return fmt.Errorf("subject is required")
	}

	if body == "" {
		es.logger.Error("Body is empty")
		return fmt.Errorf("body is required")
	}

	// Basic email validation
	if !strings.Contains(recipient, "@") || !strings.Contains(recipient, ".") {
		es.logger.Error("Invalid email format", zap.String("email", recipient))
		return fmt.Errorf("invalid email format: %s", recipient)
	}

	return nil
}

// buildMessage tạo email message theo RFC 5322
func (es *EmailSender) buildMessage(recipient, subject, body string) ([]byte, error) {
	// Headers
	headers := []string{
		fmt.Sprintf("From: %s <%s>", es.config.FromName, es.config.FromEmail),
		fmt.Sprintf("To: %s", recipient),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
	}

	// Add Reply-To if configured
	if es.config.ReplyTo != "" {
		headers = append(headers, fmt.Sprintf("Reply-To: %s", es.config.ReplyTo))
	}

	// Add Message-ID
	messageID := fmt.Sprintf("<%d.%s@%s>",
		time.Now().Unix(),
		generateRandomString(10),
		strings.Split(es.config.FromEmail, "@")[1])
	headers = append(headers, fmt.Sprintf("Message-ID: %s", messageID))

	// Build complete message
	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	return []byte(message), nil
}

// sendViaSMTP gửi email qua SMTP
func (es *EmailSender) sendViaSMTP(recipient string, message []byte) error {
	// Create SMTP auth
	auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)

	// Build server address
	addr := fmt.Sprintf("%s:%s", es.config.SMTPHost, es.config.SMTPPort)

	// Send email
	return smtp.SendMail(addr, auth, es.config.FromEmail, []string{recipient}, message)
}

// SendBulkAsync - Async version của bulk send
func (es *EmailSender) SendBulkAsync(emails []string, subject, body string) <-chan BulkEmailResult {
	resultChan := make(chan BulkEmailResult, 1)

	go func() {
		defer close(resultChan)

		result := BulkEmailResult{
			TotalEmails: len(emails),
			StartTime:   time.Now(),
		}

		for _, email := range emails {
			if err := es.Send(email, subject, body); err != nil {
				result.FailedEmails = append(result.FailedEmails, EmailFailure{
					Email: email,
					Error: err.Error(),
				})
			} else {
				result.SuccessCount++
			}

			// Add delay
			time.Sleep(100 * time.Millisecond)
		}

		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		resultChan <- result
	}()

	return resultChan
}

// TestConnection kiểm tra kết nối SMTP
func (es *EmailSender) TestConnection() error {
	es.logger.Info("Testing SMTP connection",
		zap.String("host", es.config.SMTPHost),
		zap.String("port", es.config.SMTPPort))

	auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)
	addr := fmt.Sprintf("%s:%s", es.config.SMTPHost, es.config.SMTPPort)

	// Try to connect and authenticate
	client, err := smtp.Dial(addr)
	if err != nil {
		es.logger.Error("Failed to dial SMTP server", zap.Error(err))
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Test authentication
	if err := client.Auth(auth); err != nil {
		es.logger.Error("SMTP authentication failed", zap.Error(err))
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	es.logger.Info("SMTP connection test successful")
	return nil
}

// GetConfig trả về config hiện tại (cho debugging)
func (es *EmailSender) GetConfig() EmailConfig {
	// Return copy để avoid modification
	return es.config
}

// Supporting types
type BulkEmailResult struct {
	TotalEmails  int
	SuccessCount int
	FailedEmails []EmailFailure
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
}

type EmailFailure struct {
	Email string
	Error string
}

// generateRandomString tạo random string cho Message-ID
func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}
