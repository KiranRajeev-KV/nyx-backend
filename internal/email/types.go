package email

import "context"

// IEmailService defines the interface for email operations
type IEmailService interface {
	// SendOTP sends an OTP email to the specified recipient
	SendOTP(ctx context.Context, to, otp string) error

	// SendPasswordReset sends a password reset OTP email to the specified recipient
	SendPasswordReset(ctx context.Context, to, otp string) error

	// IsEnabled returns whether email sending is enabled
	IsEnabled() bool
}

// EmailData represents the structure for email template data
type EmailData struct {
	To      string
	Subject string
	Body    string
	OTP     string
	Name    string
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	Enabled  bool
}
