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
	}
}
