package email

import (
	"context"
	"testing"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewEmailService(t *testing.T) {
	config := &cmd.EnvConfig{
		EmailEnabled:      true,
		EmailSMTPHost:     "smtp.gmail.com",
		EmailSMTPPort:     587,
		EmailFromEmail:    "test@gmail.com",
		EmailFromPassword: "password",
		EmailFromName:     "Test",
		Environment:       "TEST",
	}

	// Initialize logger for test
	log, err := logger.InitLogger("TEST")
	assert.NoError(t, err)

	service := NewEmailService(config, log)

	assert.NotNil(t, service)
	assert.True(t, service.IsEnabled())
}

func TestNewEmailService_Disabled(t *testing.T) {
	config := &cmd.EnvConfig{
		EmailEnabled: false,
		Environment:  "TEST",
	}

	// Initialize logger for test
	log, err := logger.InitLogger("TEST")
	assert.NoError(t, err)

	service := NewEmailService(config, log)

	assert.NotNil(t, service)
	assert.False(t, service.IsEnabled())
}

func TestEmailService_SendOTP_Disabled(t *testing.T) {
	// Create mock logger
	log, err := logger.InitLogger("TEST")
	assert.NoError(t, err)

	config := &cmd.EnvConfig{
		EmailEnabled: false,
		Environment:  "TEST",
	}

	service := NewEmailService(config, log)

	// Should not return error when disabled
	err = service.SendOTP(context.Background(), "test@example.com", "123456")
	assert.NoError(t, err)
}

func TestEmailService_SendPasswordReset_Disabled(t *testing.T) {
	// Create mock logger
	log, err := logger.InitLogger("TEST")
	assert.NoError(t, err)

	config := &cmd.EnvConfig{
		EmailEnabled: false,
		Environment:  "TEST",
	}

	service := NewEmailService(config, log)

	// Should not return error when disabled
	err = service.SendPasswordReset(context.Background(), "test@example.com", "123456")
	assert.NoError(t, err)
}

func TestEmailService_maskEmail(t *testing.T) {
	config := &cmd.EnvConfig{Environment: "PROD"}
	log, _ := logger.InitLogger("TEST")
	service := NewEmailService(config, log)

	tests := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "us**@example.com"},
		{"ab@domain.com", "ab@domain.com"},
		{"a@b.com", "*@b.com"},
		{"invalid-email", "***@***"},
		{"", "***@***"},
	}

	for _, test := range tests {
		result := service.maskEmail(test.input)
		assert.Equal(t, test.expected, result, "Failed for input: %s", test.input)
	}
}

func TestEmailService_maskEmail_DEV(t *testing.T) {
	config := &cmd.EnvConfig{Environment: "DEV"}
	log, _ := logger.InitLogger("TEST")
	service := NewEmailService(config, log)

	// In DEV mode, should return original email for short emails
	email := "ab@test.com"
	result := service.maskEmail(email)
	assert.Equal(t, email, result, "In DEV mode, short emails should not be masked")
}

// MockEmailService Tests

func TestMockEmailService_NewMockEmailService(t *testing.T) {
	service := NewMockEmailService(true)

	assert.NotNil(t, service)
	assert.True(t, service.IsEnabled())

	// Test initial state
	otp, exists := service.GetSentOTP("test@example.com")
	assert.Empty(t, otp)
	assert.False(t, exists)

	reset, exists := service.GetSentPasswordReset("test@example.com")
	assert.Empty(t, reset)
	assert.False(t, exists)

	attempts := service.GetSendAttempts()
	assert.Empty(t, attempts)
}

func TestMockEmailService_NewMockEmailService_Disabled(t *testing.T) {
	service := NewMockEmailService(false)

	assert.NotNil(t, service)
	assert.False(t, service.IsEnabled())
}

func TestMockEmailService_SendOTP_Enabled(t *testing.T) {
	service := NewMockEmailService(true)

	err := service.SendOTP(context.Background(), "test@example.com", "123456")
	assert.NoError(t, err)

	// Check OTP was stored
	otp, exists := service.GetSentOTP("test@example.com")
	assert.Equal(t, "123456", otp)
	assert.True(t, exists)

	// Check attempts were recorded
	attempts := service.GetSendAttempts()
	assert.Len(t, attempts, 1)
	assert.Equal(t, "test@example.com", attempts[0].To)
	assert.Equal(t, "123456", attempts[0].OTP)
	assert.Equal(t, "OTP", attempts[0].Type)
}

func TestMockEmailService_SendOTP_Disabled(t *testing.T) {
	service := NewMockEmailService(false)

	err := service.SendOTP(context.Background(), "test@example.com", "123456")
	assert.NoError(t, err)

	// Check OTP was not stored
	otp, exists := service.GetSentOTP("test@example.com")
	assert.Empty(t, otp)
	assert.False(t, exists)

	// Check no attempts were recorded
	attempts := service.GetSendAttempts()
	assert.Empty(t, attempts)
}

func TestMockEmailService_SendPasswordReset_Enabled(t *testing.T) {
	service := NewMockEmailService(true)

	err := service.SendPasswordReset(context.Background(), "test@example.com", "654321")
	assert.NoError(t, err)

	// Check reset OTP was stored
	reset, exists := service.GetSentPasswordReset("test@example.com")
	assert.Equal(t, "654321", reset)
	assert.True(t, exists)

	// Check attempts were recorded
	attempts := service.GetSendAttempts()
	assert.Len(t, attempts, 1)
	assert.Equal(t, "test@example.com", attempts[0].To)
	assert.Equal(t, "654321", attempts[0].OTP)
	assert.Equal(t, "RESET", attempts[0].Type)
}

func TestMockEmailService_SendPasswordReset_Disabled(t *testing.T) {
	service := NewMockEmailService(false)

	err := service.SendPasswordReset(context.Background(), "test@example.com", "654321")
	assert.NoError(t, err)

	// Check reset OTP was not stored
	reset, exists := service.GetSentPasswordReset("test@example.com")
	assert.Empty(t, reset)
	assert.False(t, exists)

	// Check no attempts were recorded
	attempts := service.GetSendAttempts()
	assert.Empty(t, attempts)
}

func TestMockEmailService_SetShouldError(t *testing.T) {
	service := NewMockEmailService(true)

	// Enable error mode
	service.SetShouldError(true)

	err := service.SendOTP(context.Background(), "test@example.com", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Mock email service forced error")

	// Disable error mode
	service.SetShouldError(false)

	err = service.SendOTP(context.Background(), "test@example.com", "123456")
	assert.NoError(t, err)
}

func TestMockEmailService_SetDelay(t *testing.T) {
	service := NewMockEmailService(true)

	// Set a small delay
	service.SetDelay(10 * time.Millisecond)

	start := time.Now()
	err := service.SendOTP(context.Background(), "test@example.com", "123456")
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond)
	assert.Less(t, duration, 50*time.Millisecond) // Allow some tolerance
}

func TestMockEmailService_GetAttemptsByEmail(t *testing.T) {
	service := NewMockEmailService(true)

	// Send multiple emails to different addresses
	service.SendOTP(context.Background(), "user1@example.com", "111111")
	service.SendPasswordReset(context.Background(), "user2@example.com", "222222")
	service.SendOTP(context.Background(), "user1@example.com", "333333")

	// Get attempts for user1
	user1Attempts := service.GetAttemptsByEmail("user1@example.com")
	assert.Len(t, user1Attempts, 2)

	// Get attempts for user2
	user2Attempts := service.GetAttemptsByEmail("user2@example.com")
	assert.Len(t, user2Attempts, 1)

	// Get attempts for non-existent user
	user3Attempts := service.GetAttemptsByEmail("user3@example.com")
	assert.Empty(t, user3Attempts)
}

func TestMockEmailService_Clear(t *testing.T) {
	service := NewMockEmailService(true)

	// Send some emails
	service.SendOTP(context.Background(), "test@example.com", "123456")
	service.SendPasswordReset(context.Background(), "test@example.com", "654321")

	// Verify data exists
	assert.NotEmpty(t, service.GetSendAttempts())

	// Clear all data
	service.Clear()

	// Verify data is cleared
	attempts := service.GetSendAttempts()
	assert.Empty(t, attempts)

	otp, exists := service.GetSentOTP("test@example.com")
	assert.Empty(t, otp)
	assert.False(t, exists)

	reset, exists := service.GetSentPasswordReset("test@example.com")
	assert.Empty(t, reset)
	assert.False(t, exists)
}

func TestEmailError_Error(t *testing.T) {
	err := &EmailError{Message: "Test error"}
	assert.Equal(t, "Test error", err.Error())
}

func TestSendAttempt_Structure(t *testing.T) {
	now := time.Now()
	attempt := SendAttempt{
		To:      "test@example.com",
		Subject: "Test Subject",
		Body:    "Test Body",
		Time:    now,
		Type:    "OTP",
		OTP:     "123456",
	}

	assert.Equal(t, "test@example.com", attempt.To)
	assert.Equal(t, "Test Subject", attempt.Subject)
	assert.Equal(t, "Test Body", attempt.Body)
	assert.Equal(t, now, attempt.Time)
	assert.Equal(t, "OTP", attempt.Type)
	assert.Equal(t, "123456", attempt.OTP)
}

func TestMockEmailService_ConcurrentAccess(t *testing.T) {
	service := NewMockEmailService(true)

	// Test concurrent access
	done := make(chan bool, 3)

	// Goroutine 1: Send OTP
	go func() {
		service.SendOTP(context.Background(), "test1@example.com", "111111")
		done <- true
	}()

	// Goroutine 2: Send Password Reset
	go func() {
		service.SendPasswordReset(context.Background(), "test2@example.com", "222222")
		done <- true
	}()

	// Goroutine 3: Read attempts
	go func() {
		for i := 0; i < 10; i++ {
			service.GetSendAttempts()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify all emails were sent
	attempts := service.GetSendAttempts()
	assert.Len(t, attempts, 2)

	otp, exists := service.GetSentOTP("test1@example.com")
	assert.Equal(t, "111111", otp)
	assert.True(t, exists)

	reset, exists := service.GetSentPasswordReset("test2@example.com")
	assert.Equal(t, "222222", reset)
	assert.True(t, exists)
}
