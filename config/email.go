package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"anhnq/api-core/pkg/email"
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
	config := &EmailConfig{
		// Default values for MailHog
		SMTPHost:     "localhost",
		SMTPPort:     1025,
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "noreply@apicore.com",
		FromName:     "ApiCore",
		UseTLS:       false,
	}

	// Load from environment variables
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.SMTPHost = smtpHost
	}

	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		if port, err := strconv.Atoi(smtpPort); err == nil {
			config.SMTPPort = port
		}
	}

	if smtpUsername := os.Getenv("SMTP_USERNAME"); smtpUsername != "" {
		config.SMTPUsername = smtpUsername
	}

	if smtpPassword := os.Getenv("SMTP_PASSWORD"); smtpPassword != "" {
		config.SMTPPassword = smtpPassword
	}

	if fromEmail := os.Getenv("EMAIL_FROM"); fromEmail != "" {
		config.FromEmail = fromEmail
	}

	if fromName := os.Getenv("EMAIL_FROM_NAME"); fromName != "" {
		config.FromName = fromName
	}

	if useTLS := os.Getenv("SMTP_USE_TLS"); useTLS != "" {
		config.UseTLS = strings.ToLower(useTLS) == "true"
	}

	return config
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
