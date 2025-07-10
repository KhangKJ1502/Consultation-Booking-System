package email

import "cbs_backend/global"

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string
	ReplyTo      string // Thêm ReplyTo field
	Timeout      int    // Timeout in seconds
	MaxRetries   int    // Max retry attempts
	TLSEnabled   bool   // Enable TLS
}

func LoadEmailConfig() EmailConfig {
	config := global.ConfigConection.SMTPCF
	return EmailConfig{
		SMTPHost:     config.SmtpHost,
		SMTPPort:     config.SmtpPort,
		SMTPUsername: config.SmtpUsername,
		SMTPPassword: config.SmtpPassword,
		FromEmail:    config.FromEmail,
		FromName:     config.FromName,
		BaseURL:      config.BaseURL,
		ReplyTo:      config.ReplyTo, // Nếu có trong global config
		Timeout:      30,
		MaxRetries:   3,
		TLSEnabled:   true,
	}
}
