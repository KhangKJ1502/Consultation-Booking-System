package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Constants
const (
	EmailSendDelay = 100 * time.Millisecond
)

// Structs
type NotificationService struct {
	db *gorm.DB
}

type EmailBatch struct {
	Recipients []string               `json:"recipients"`
	Subject    string                 `json:"subject"`
	Body       string                 `json:"body"`
	Template   string                 `json:"template"`
	Data       map[string]interface{} `json:"data"`
}

type EmailResult struct {
	Recipient string
	Success   bool
	Error     error
}

// Constructor
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// Main processing method
func (ns *NotificationService) ProcessEmailBatch(payload interface{}) error {
	log.Println("📧 Processing email batch...")

	batch, err := ns.parseEmailBatch(payload)
	if err != nil {
		return fmt.Errorf("failed to parse email batch: %w", err)
	}

	if len(batch.Recipients) == 0 {
		log.Println("⚠️ No recipients in email batch")
		return nil
	}

	results := ns.sendEmailBatch(batch)
	ns.logBatchResults(results)

	log.Printf("✅ Email batch processed: %d recipients", len(batch.Recipients))
	return nil
}

// Helper methods
func (ns *NotificationService) parseEmailBatch(payload interface{}) (*EmailBatch, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var batch EmailBatch
	if err := json.Unmarshal(payloadBytes, &batch); err != nil {
		return nil, fmt.Errorf("failed to unmarshal email batch: %w", err)
	}

	return &batch, nil
}

func (ns *NotificationService) sendEmailBatch(batch *EmailBatch) []EmailResult {
	results := make([]EmailResult, 0, len(batch.Recipients))

	for _, recipient := range batch.Recipients {
		result := EmailResult{
			Recipient: recipient,
			Success:   true,
		}

		if err := ns.sendEmail(recipient, batch.Subject, batch.Body, batch.Template, batch.Data); err != nil {
			result.Success = false
			result.Error = err
		}

		results = append(results, result)
	}

	return results
}

func (ns *NotificationService) logBatchResults(results []EmailResult) {
	successCount := 0
	for _, result := range results {
		if result.Success {
			log.Printf("✅ Email sent to %s", result.Recipient)
			successCount++
		} else {
			log.Printf("❌ Failed to send email to %s: %v", result.Recipient, result.Error)
		}
	}

	log.Printf("📊 Batch summary: %d/%d emails sent successfully", successCount, len(results))
}

// Email sending implementation
func (ns *NotificationService) sendEmail(recipient, subject, body, template string, data map[string]interface{}) error {
	if recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}

	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}

	log.Printf("📧 Sending email to %s: %s", recipient, subject)

	// Simulate email sending - replace with actual email service
	if err := ns.deliverEmail(recipient, subject, body, template, data); err != nil {
		return fmt.Errorf("failed to deliver email: %w", err)
	}

	return nil
}

func (ns *NotificationService) deliverEmail(recipient, subject, body, template string, data map[string]interface{}) error {
	// Here you would integrate with actual email service such as:
	// - SendGrid: sendgrid.NewSendClient(apiKey).Send(email)
	// - AWS SES: ses.SendEmail(input)
	// - SMTP server: smtp.SendMail(addr, auth, from, to, msg)
	// - Mailgun: mg.Send(message)
	// - etc.

	// For demonstration, simulate email delivery delay
	time.Sleep(EmailSendDelay)

	// Simulate potential failure (uncomment to test error handling)
	// if strings.Contains(recipient, "fail") {
	//     return fmt.Errorf("simulated email delivery failure")
	// }

	return nil
}

// Additional utility methods
func (ns *NotificationService) ValidateEmailBatch(batch *EmailBatch) error {
	if len(batch.Recipients) == 0 {
		return fmt.Errorf("recipients list cannot be empty")
	}

	if batch.Subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}

	if batch.Body == "" && batch.Template == "" {
		return fmt.Errorf("either body or template must be provided")
	}

	for i, recipient := range batch.Recipients {
		if recipient == "" {
			return fmt.Errorf("recipient at index %d cannot be empty", i)
		}
		// Add email format validation if needed
		// if !isValidEmail(recipient) {
		//     return fmt.Errorf("invalid email format: %s", recipient)
		// }
	}

	return nil
}

func (ns *NotificationService) GetEmailStats() map[string]interface{} {
	// This could return statistics from database
	// For now, return placeholder data
	return map[string]interface{}{
		"total_sent":   0,
		"total_failed": 0,
		"last_sent":    time.Now().Format(time.RFC3339),
	}
}

// Example of how to extend for other notification types
func (ns *NotificationService) ProcessSMSBatch(payload interface{}) error {
	log.Println("📱 Processing SMS batch...")
	// Similar implementation for SMS
	return nil
}

func (ns *NotificationService) ProcessPushNotificationBatch(payload interface{}) error {
	log.Println("📲 Processing push notification batch...")
	// Similar implementation for push notifications
	return nil
}
