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
