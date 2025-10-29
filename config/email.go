package config

import (
	"fmt"

	"api-core/pkg/email"
	"api-core/pkg/utils"
)

// EmailConfig cấu hình cho email service
type EmailConfig struct {
	SMTPHost     string // SMTP server host
	SMTPPort     int    // SMTP server port
	SMTPUsername string // SMTP username
	SMTPPassword string // SMTP password
	FromEmail    string // From email address
	FromName     string // From name
	UseTLS       bool   // Use TLS encryption
}

// LoadEmailConfig load email config từ environment variables
func LoadEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTPHost:     utils.GetEnv("SMTP_HOST", "localhost"),
		SMTPPort:     utils.GetEnvInt("SMTP_PORT", 1025),
		SMTPUsername: utils.GetEnv("SMTP_USERNAME", ""),
		SMTPPassword: utils.GetEnv("SMTP_PASSWORD", ""),
		FromEmail:    utils.GetEnv("EMAIL_FROM", "noreply@apicore.com"),
		FromName:     utils.GetEnv("EMAIL_FROM_NAME", "ApiCore"),
		UseTLS:       utils.GetEnvBool("SMTP_USE_TLS", false),
	}
}

// ValidateEmailConfig kiểm tra email config có hợp lệ không
func (c *EmailConfig) Validate() error {
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}

	if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
		return fmt.Errorf("SMTP port must be between 1 and 65535")
	}

	if c.FromEmail == "" {
		return fmt.Errorf("from email is required")
	}

	return nil
}

// ToEmailConfig convert sang email.EmailConfig
func (c *EmailConfig) ToEmailConfig() email.EmailConfig {
	return email.EmailConfig{
		SMTPHost:     c.SMTPHost,
		SMTPPort:     c.SMTPPort,
		SMTPUsername: c.SMTPUsername,
		SMTPPassword: c.SMTPPassword,
		FromEmail:    c.FromEmail,
		FromName:     c.FromName,
		UseTLS:       c.UseTLS,
	}
}
