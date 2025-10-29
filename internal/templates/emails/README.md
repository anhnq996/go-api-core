# Email Templates

Thư mục này chứa các email templates cho ApiCore application.

## Templates Available

### 1. Welcome Email (`welcome.html`)

- **Mục đích**: Gửi email chào mừng khi user đăng ký tài khoản mới
- **Variables**:
  - `{{.Name}}` - Tên của user
  - `{{.Email}}` - Email của user
  - `{{.URL}}` - URL để user bắt đầu sử dụng

### 2. Password Reset Email (`password_reset.html`)

- **Mục đích**: Gửi email reset password khi user quên mật khẩu
- **Variables**:
  - `{{.Name}}` - Tên của user
  - `{{.Email}}` - Email của user
  - `{{.ResetURL}}` - URL để reset password

### 3. Notification Email (`notification.html`)

- **Mục đích**: Gửi email thông báo chung
- **Variables**:
  - `{{.Name}}` - Tên của user
  - `{{.Email}}` - Email của user
  - `{{.Subject}}` - Tiêu đề email
  - `{{.Content}}` - Nội dung email

### 4. Email Verification (`verification.html`)

- **Mục đích**: Gửi email xác thực email address
- **Variables**:
  - `{{.Name}}` - Tên của user
  - `{{.Email}}` - Email của user
  - `{{.VerificationURL}}` - URL để verify email

## Usage

### Trong Go Code

```go
package main

import (
    "api-core/pkg/email"
)

func main() {
    // Tạo email manager
    manager := email.NewEmailManager(config)

    // Tạo email message
    message := &email.EmailMessage{
        To:      []string{"user@example.com"},
        Subject: "Welcome to ApiCore",
    }

    // Data để truyền vào template
    data := struct {
        Name  string
        Email string
        URL   string
    }{
        Name:  "John Doe",
        Email: "user@example.com",
        URL:   "https://apicore.com/dashboard",
    }

    // Gửi email với template
    err := manager.SendTemplate(message, "internal/templates/emails/welcome.html", data)
    if err != nil {
        log.Printf("Failed to send email: %v", err)
    }
}
```

### Variables Structure

```go
// Welcome Email Data
type WelcomeData struct {
    Name  string
    Email string
    URL   string
}

// Password Reset Email Data
type PasswordResetData struct {
    Name      string
    Email     string
    ResetURL  string
}

// Notification Email Data
type NotificationData struct {
    Name    string
    Email   string
    Subject string
    Content string
}

// Verification Email Data
type VerificationData struct {
    Name            string
    Email           string
    VerificationURL string
}
```

## Template Syntax

Templates sử dụng Go template syntax:

- `{{.VariableName}}` - In ra giá trị của variable
- `{{if .Condition}}...{{end}}` - Conditional rendering
- `{{range .Items}}...{{end}}` - Loop rendering
- `{{.VariableName | html}}` - Escape HTML
- `{{.VariableName | urlquery}}` - URL encode

## Styling

Templates sử dụng inline CSS để đảm bảo tương thích với email clients:

- Responsive design
- Fallback fonts
- Inline styles
- Table-based layout for better compatibility

## Testing

Để test templates:

1. Start MailHog: `mailhog`
2. Access MailHog UI: http://localhost:8025
3. Run email tests: `go test ./pkg/email`
4. Check emails trong MailHog UI

## Best Practices

1. **Responsive Design**: Sử dụng responsive CSS
2. **Fallback Content**: Cung cấp text version
3. **Accessibility**: Sử dụng alt text cho images
4. **Testing**: Test với nhiều email clients
5. **Performance**: Optimize images và CSS
6. **Security**: Validate và escape user input

## Email Clients Compatibility

Templates được thiết kế để tương thích với:

- Gmail
- Outlook
- Apple Mail
- Yahoo Mail
- Thunderbird
- Mobile email clients

## Customization

Để customize templates:

1. Edit HTML files trong thư mục này
2. Update CSS styles
3. Modify template variables
4. Test với MailHog
5. Deploy changes

## Support

Nếu có vấn đề với templates:

1. Check template syntax
2. Verify variable names
3. Test với MailHog
4. Check email client compatibility
5. Review CSS styles
