# Email Package

Package email cung cấp chức năng gửi email với hỗ trợ SMTP và MailHog để test.

## Features

- ✅ **SMTP Support**: Hỗ trợ gửi email qua SMTP
- ✅ **MailHog Integration**: Tích hợp với MailHog để test
- ✅ **HTML & Text**: Hỗ trợ cả HTML và text email
- ✅ **Attachments**: Hỗ trợ file đính kèm
- ✅ **Templates**: Hỗ trợ email templates
- ✅ **Multiple Recipients**: Hỗ trợ To, CC, BCC
- ✅ **TLS Support**: Hỗ trợ TLS encryption

## Installation

```bash
go get gopkg.in/gomail.v2
```

## Configuration

### Environment Variables

```env
# Email Configuration
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_USE_TLS=false
EMAIL_FROM=noreply@apicore.com
EMAIL_FROM_NAME=ApiCore
```

### MailHog Setup

1. **Install MailHog**:

```bash
# Windows
go install github.com/mailhog/MailHog@latest

# macOS
brew install mailhog

# Linux
wget https://github.com/mailhog/MailHog/releases/download/v1.0.1/MailHog_linux_amd64
chmod +x MailHog_linux_amd64
sudo mv MailHog_linux_amd64 /usr/local/bin/mailhog
```

2. **Start MailHog**:

```bash
mailhog
```

3. **Access MailHog UI**: http://localhost:8025

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "api-core/pkg/email"
)

func main() {
    // Tạo email config
    config := email.EmailConfig{
        SMTPHost:     "localhost",
        SMTPPort:     1025,
        SMTPUsername: "",
        SMTPPassword: "",
        FromEmail:    "noreply@apicore.com",
        FromName:     "ApiCore",
        UseTLS:       false,
    }

    // Tạo email manager
    manager := email.NewEmailManager(config)

    // Tạo email message
    message := &email.EmailMessage{
        To:      []string{"user@example.com"},
        Subject: "Test Email",
        Body:    "<h1>Hello World!</h1><p>This is a test email.</p>",
        TextBody: "Hello World!\n\nThis is a test email.",
    }

    // Gửi email
    if err := manager.Send(message); err != nil {
        log.Printf("Failed to send email: %v", err)
        return
    }

    log.Println("Email sent successfully!")
}
```

### Welcome Email

```go
// Gửi email chào mừng
if err := manager.SendWelcomeEmail("user@example.com", "John Doe"); err != nil {
    log.Printf("Failed to send welcome email: %v", err)
    return
}
```

### Password Reset Email

```go
// Gửi email reset password
resetToken := "abc123def456"
if err := manager.SendPasswordResetEmail("user@example.com", resetToken); err != nil {
    log.Printf("Failed to send password reset email: %v", err)
    return
}
```

### Notification Email

```go
// Gửi email thông báo
recipients := []string{"user1@example.com", "user2@example.com"}
subject := "System Maintenance"
content := "System will be under maintenance from 2:00 AM to 4:00 AM tomorrow."

if err := manager.SendNotificationEmail(recipients, subject, content); err != nil {
    log.Printf("Failed to send notification email: %v", err)
    return
}
```

### Email with Attachment

```go
// Tạo email message với file đính kèm
message := &email.EmailMessage{
    To:      []string{"user@example.com"},
    Subject: "Email with Attachment",
    Body:    "<h1>Hello!</h1><p>Please find the attached file.</p>",
    TextBody: "Hello!\n\nPlease find the attached file.",
    Attachments: []email.Attachment{
        {
            Filename: "report.pdf",
            Content:  []byte("This is a fake PDF content"),
            MimeType: "application/pdf",
        },
    },
}

// Gửi email
if err := manager.Send(message); err != nil {
    log.Printf("Failed to send email with attachment: %v", err)
    return
}
```

### Email with Template

```go
// Gửi email với template
message := &email.EmailMessage{
    To:      []string{"user@example.com"},
    Subject: "Welcome Email",
}

data := struct {
    Name string
    URL  string
}{
    Name: "John Doe",
    URL:  "https://apicore.com",
}

if err := manager.SendTemplate(message, "templates/welcome.html", data); err != nil {
    log.Printf("Failed to send template email: %v", err)
    return
}
```

## Email Templates

### Welcome Template (templates/welcome.html)

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>Welcome</title>
  </head>
  <body>
    <h1>Welcome {{.Name}}!</h1>
    <p>Thank you for joining ApiCore.</p>
    <p>Visit us at: <a href="{{.URL}}">{{.URL}}</a></p>
    <br />
    <p>Best regards,<br />Team ApiCore</p>
  </body>
</html>
```

## Production Configuration

### Gmail SMTP

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_TLS=true
EMAIL_FROM=your-email@gmail.com
EMAIL_FROM_NAME=Your App Name
```

### SendGrid

```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
SMTP_USE_TLS=true
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Your App Name
```

### AWS SES

```env
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=your-ses-username
SMTP_PASSWORD=your-ses-password
SMTP_USE_TLS=true
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=Your App Name
```

## Best Practices

1. **Use Templates**: Sử dụng templates cho email để dễ maintain
2. **Error Handling**: Luôn handle errors khi gửi email
3. **Rate Limiting**: Implement rate limiting để tránh spam
4. **Validation**: Validate email addresses trước khi gửi
5. **Logging**: Log email sending activities
6. **Testing**: Test với MailHog trong development
7. **Security**: Sử dụng TLS trong production
8. **Monitoring**: Monitor email delivery rates

## Troubleshooting

### Common Issues

1. **Connection Refused**: Kiểm tra SMTP server có đang chạy không
2. **Authentication Failed**: Kiểm tra username/password
3. **TLS Errors**: Kiểm tra TLS configuration
4. **Email Not Delivered**: Kiểm tra spam folder
5. **Template Errors**: Kiểm tra template syntax

### Debug Mode

```go
// Enable debug mode
config.Debug = true
manager := email.NewEmailManager(config)
```

## License

MIT License
