# Email Service Documentation

## Overview

The Nyx backend includes a comprehensive email service for sending OTPs and password reset emails via SMTP. The service supports configuration through environment variables and includes a toggle to enable/disable email sending.

## Configuration

Add the following to your `.env` file:

```bash
# Email Configuration
EMAIL_ENABLE=true                       # Toggle email sending on/off
EMAIL_SMTP_HOST=smtp.gmail.com           # SMTP server host
EMAIL_SMTP_PORT=587                      # SMTP server port
EMAIL_FROM_EMAIL=your-email@gmail.com       # From email address
EMAIL_FROM_PASSWORD=your-app-password       # SMTP password (use app password for Gmail)
EMAIL_FROM_NAME="Nyx System"            # Display name for emails
```

### Gmail Setup

1. Enable 2-factor authentication on your Google Account
2. Generate an App Password:
   - Go to Google Account settings → Security → 2-Step Verification → App passwords
   - Select "Mail" app and generate password
   - Use this app password instead of your regular password

## Features

### ✅ Implemented Features

- **Toggle Support**: `EMAIL_ENABLE` flag controls all email sending
- **Production Safe Logging**: OTP values masked in production logs
- **SMTP Integration**: Uses Go's standard `net/smtp` package
- **STARTTLS Support**: Secure email transmission
- **Interface Design**: Easy to mock for testing
- **Comprehensive Error Handling**: Graceful failure modes

### 📧 Email Types Supported

1. **OTP Verification Emails**: Sent during user registration and resend
2. **Password Reset Emails**: Sent during password reset flow

### 🔒 Security Features

- **Email Masking in Production**: `user@example.com` → `us**@example.com`
- **Development Mode**: Full email visibility for debugging
- **Error Logging**: Detailed error logging without exposing sensitive data

## Usage

### In Application Code

```go
// Email service is automatically initialized in main.go
// EmailService is available in auth controllers

// Send OTP
err := EmailService.SendOTP(context.Background(), "user@example.com", "123456")

// Send Password Reset
err := EmailService.SendPasswordReset(context.Background(), "user@example.com", "654321")

// Check if enabled
if EmailService.IsEnabled() {
    // Send email
}
```

### For Testing

```go
// Use mock service for unit tests
mockService := email.NewMockEmailService(true)

// Send emails (records them instead of actually sending)
err := mockService.SendOTP(context.Background(), "test@example.com", "123456")

// Verify emails were "sent"
otp, exists := mockService.GetSentOTP("test@example.com")
assert.True(t, exists)
assert.Equal(t, "123456", otp)

// Get send attempts
attempts := mockService.GetSendAttempts()
assert.Len(t, attempts, 1)
```

## Environment Behavior

### Development (`EMAIL_ENABLE=true`)
- Emails are actually sent via SMTP
- Full email addresses visible in logs
- Detailed error messages for debugging

### Production (`EMAIL_ENABLE=true`)
- Emails are actually sent via SMTP
- Email addresses masked in logs: `user@example.com` → `us**@example.com`
- OTP values not logged

### Testing (`EMAIL_ENABLE=false`)
- Emails are NOT sent
- `EmailService.IsEnabled()` returns `false`
- No SMTP connections attempted
- Application continues normally (useful for CI/CD)

## Error Handling

### SMTP Connection Errors
- Logged with masked email addresses in production
- Application continues gracefully
- User can retry operations

### Configuration Errors
- SMTP port validation in configuration
- Graceful fallback when email service is disabled
- Detailed error messages for misconfiguration

## Testing

Run the email service tests:

```bash
go test ./internal/email/... -v
```

Run all application tests:

```bash
go test ./... -v
```

## Troubleshooting

### Gmail Issues
1. **"Authentication failed"** - Use an App Password, not your regular password
2. **"Could not connect"** - Check SMTP host and port settings
3. **"535 5.7.8 Username and Password not accepted"** - Enable 2FA and use App Password

### Testing Issues
1. **Emails not being "sent"** - Check `EMAIL_ENABLE=true`
2. **Mock service not working** - Ensure you're using the correct import path

### Production Issues
1. **Emails not arriving** - Check spam folders, SMTP credentials
2. **Slow response times** - Consider SMTP server timeouts
3. **High failure rates** - Monitor rate limiting and server health

## File Structure

```
internal/email/
├── service.go          # Main SMTP implementation
├── types.go           # Interface and type definitions
├── mock_service.go    # Mock implementation for testing
└── service_test.go    # Comprehensive unit tests
```