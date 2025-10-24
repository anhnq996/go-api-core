package email

import (
	"bytes"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"
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

// EmailMessage đại diện cho một email message
type EmailMessage struct {
	To          []string     // Danh sách người nhận
	CC          []string     // Danh sách CC
	BCC         []string     // Danh sách BCC
	Subject     string       // Tiêu đề email
	Body        string       // Nội dung email (HTML)
	TextBody    string       // Nội dung email (Text)
	Attachments []Attachment // Danh sách file đính kèm
}

// Attachment đại diện cho file đính kèm
type Attachment struct {
	Filename string // Tên file
	Content  []byte // Nội dung file
	MimeType string // MIME type
}

// EmailService interface cho email service
type EmailService interface {
	Send(message *EmailMessage) error
	SendTemplate(message *EmailMessage, templatePath string, data interface{}) error
}

// emailService implementation của EmailService
type emailService struct {
	config EmailConfig
	dialer *gomail.Dialer
}

// NewEmailService tạo mới email service
func NewEmailService(config EmailConfig) EmailService {
	dialer := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)
	if config.UseTLS {
		dialer.TLSConfig = nil // Use default TLS config
	}

	return &emailService{
		config: config,
		dialer: dialer,
	}
}

// Send gửi email message
func (e *emailService) Send(message *EmailMessage) error {
	m := gomail.NewMessage()

	// Set sender
	if e.config.FromName != "" {
		m.SetHeader("From", fmt.Sprintf("%s <%s>", e.config.FromName, e.config.FromEmail))
	} else {
		m.SetHeader("From", e.config.FromEmail)
	}

	// Set recipients
	if len(message.To) > 0 {
		m.SetHeader("To", message.To...)
	}
	if len(message.CC) > 0 {
		m.SetHeader("Cc", message.CC...)
	}
	if len(message.BCC) > 0 {
		m.SetHeader("Bcc", message.BCC...)
	}

	// Set subject
	m.SetHeader("Subject", message.Subject)

	// Set body
	if message.TextBody != "" && message.Body != "" {
		// Both HTML and text versions
		m.SetBody("text/plain", message.TextBody)
		m.AddAlternative("text/html", message.Body)
	} else if message.Body != "" {
		// HTML only
		m.SetBody("text/html", message.Body)
	} else if message.TextBody != "" {
		// Text only
		m.SetBody("text/plain", message.TextBody)
	}

	// Add attachments (TODO: Fix attachment implementation)
	// for _, attachment := range message.Attachments {
	// 	m.Attach(attachment.Filename, gomail.SetCopyFunc(func(w *gomail.Message) error {
	// 		return w.SetBody("application/octet-stream", string(attachment.Content))
	// 	}))
	// }

	// Send email
	return e.dialer.DialAndSend(m)
}

// SendTemplate gửi email với template
func (e *emailService) SendTemplate(message *EmailMessage, templatePath string, data interface{}) error {
	// Load template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Set rendered body
	message.Body = body.String()

	// Send email
	return e.Send(message)
}

// EmailManager quản lý email service
type EmailManager struct {
	service EmailService
	config  EmailConfig
}

// NewEmailManager tạo mới email manager
func NewEmailManager(config EmailConfig) *EmailManager {
	return &EmailManager{
		service: NewEmailService(config),
		config:  config,
	}
}

// Send gửi email message
func (em *EmailManager) Send(message *EmailMessage) error {
	return em.service.Send(message)
}

// SendTemplate gửi email với template
func (em *EmailManager) SendTemplate(message *EmailMessage, templatePath string, data interface{}) error {
	return em.service.SendTemplate(message, templatePath, data)
}
