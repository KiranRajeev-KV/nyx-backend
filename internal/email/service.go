package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
)

// EmailService implements the IEmailService interface
type EmailService struct {
	config *cmd.EnvConfig
	logger logger.ILoggerService
}

// NewEmailService creates a new instance of EmailService
func NewEmailService(config *cmd.EnvConfig, logger logger.ILoggerService) *EmailService {
	return &EmailService{
		config: config,
		logger: logger,
	}
}

// IsEnabled returns whether email sending is enabled
func (es *EmailService) IsEnabled() bool {
	return es.config.EmailEnabled
}

// SendOTP sends an OTP email to the specified recipient
func (es *EmailService) SendOTP(ctx context.Context, to, otp string) error {
	if !es.IsEnabled() {
		es.logger.Info("[EMAIL-SKIPPED]: Email sending is disabled")
		return nil
	}

	subject := "Nyx - Email Verification"
	body := fmt.Sprintf(`Hello,

Your verification code is: %s

This code will expire in 5 minutes.

If you didn't request this code, please ignore this email.

Best regards,
Nyx Team`, otp)

	return es.sendEmail(ctx, to, subject, body)
}

// SendPasswordReset sends a password reset OTP email to the specified recipient
func (es *EmailService) SendPasswordReset(ctx context.Context, to, otp string) error {
	if !es.IsEnabled() {
		es.logger.Info("[EMAIL-SKIPPED]: Email sending is disabled")
		return nil
	}

	subject := "Nyx - Password Reset"
	body := fmt.Sprintf(`Hello,

You requested to reset your password. Your reset code is: %s

This code will expire in 10 minutes.

If you didn't request this reset, please secure your account immediately.

Best regards,
Nyx Team`, otp)

	return es.sendEmail(ctx, to, subject, body)
}

// sendEmail is the internal method that handles SMTP email sending
func (es *EmailService) sendEmail(ctx context.Context, to, subject, body string) error {
	if !es.IsEnabled() {
		return nil
	}

	// Build SMTP auth
	auth := smtp.PlainAuth("", es.config.EmailFromEmail, es.config.EmailFromPassword, es.config.EmailSMTPHost)

	// Build email message
	from := fmt.Sprintf("%s <%s>", es.config.EmailFromName, es.config.EmailFromEmail)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", es.config.EmailSMTPHost, es.config.EmailSMTPPort)

	// Send email using STARTTLS
	client, err := smtp.Dial(addr)
	if err != nil {
		es.logEmailError(ctx, to, subject, "failed to connect to SMTP server", err)
		return err
	}
	defer client.Close()

	// Start TLS if available
	tlsConfig := &tls.Config{
		ServerName: es.config.EmailSMTPHost,
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			es.logEmailError(ctx, to, subject, "failed to start TLS", err)
			return err
		}
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		es.logEmailError(ctx, to, subject, "SMTP authentication failed", err)
		return err
	}

	// Send email
	if err = client.Mail(es.config.EmailFromEmail); err != nil {
		es.logEmailError(ctx, to, subject, "failed to set sender", err)
		return err
	}

	if err = client.Rcpt(to); err != nil {
		es.logEmailError(ctx, to, subject, "failed to set recipient", err)
		return err
	}

	w, err := client.Data()
	if err != nil {
		es.logEmailError(ctx, to, subject, "failed to send data", err)
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		es.logEmailError(ctx, to, subject, "failed to write message", err)
		return err
	}

	err = w.Close()
	if err != nil {
		es.logEmailError(ctx, to, subject, "failed to close message", err)
		return err
	}

	es.logEmailSent(ctx, to, subject)
	return nil
}

// logEmailSent logs successful email sending (production-safe)
func (es *EmailService) logEmailSent(ctx context.Context, to, subject string) {
	message := fmt.Sprintf("[EMAIL-SENT]: Email sent to %s with subject '%s'", to, subject)
	if es.config.Environment == "DEV" {
		es.logger.Info(message)
	} else {
		maskedEmail := es.maskEmail(to)
		message = fmt.Sprintf("[EMAIL-SENT]: Email sent successfully to %s with subject '%s'", maskedEmail, subject)
		es.logger.Info(message)
	}
}

// logEmailError logs email sending errors (production-safe)
func (es *EmailService) logEmailError(ctx context.Context, to, subject, operation string, err error) {
	message := fmt.Sprintf("[EMAIL-ERROR]: %s to %s", operation, to)
	if es.config.Environment == "DEV" {
		es.logger.Error(message, err)
	} else {
		maskedEmail := es.maskEmail(to)
		message = fmt.Sprintf("[EMAIL-ERROR]: %s to %s", operation, maskedEmail)
		es.logger.Error(message, err)
	}
}

// maskEmail masks email addresses for production logging
func (es *EmailService) maskEmail(email string) string {
	// In DEV mode, don't mask emails to aid debugging
	if es.config.Environment == "DEV" {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) <= 1 {
		localPart = strings.Repeat("*", len(localPart))
	} else if len(localPart) == 2 {
		// Don't mask 2-character local parts
	} else {
		localPart = localPart[:2] + strings.Repeat("*", len(localPart)-2)
	}

	return fmt.Sprintf("%s@%s", localPart, domainPart)
}
