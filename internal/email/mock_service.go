package email

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockEmailService is a mock implementation of IEmailService for testing
type MockEmailService struct {
	mu           sync.RWMutex
	enabled      bool
	sentOTPs     map[string]string // email -> otp
	sentResets   map[string]string // email -> otp
	sendAttempts []SendAttempt
	delay        time.Duration // Optional delay for testing
	shouldError  bool          // Force errors for testing
}

// SendAttempt records an email send attempt
type SendAttempt struct {
	To      string
	Subject string
	Body    string
	Time    time.Time
	Type    string // "OTP" or "RESET"
	OTP     string
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService(enabled bool) *MockEmailService {
	return &MockEmailService{
		enabled:      enabled,
		sentOTPs:     make(map[string]string),
		sentResets:   make(map[string]string),
		sendAttempts: make([]SendAttempt, 0),
	}
}

// SetDelay sets a delay for email sending (useful for testing timeouts)
func (m *MockEmailService) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = delay
}

// SetShouldError forces the service to return errors (useful for testing error scenarios)
func (m *MockEmailService) SetShouldError(shouldError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
}

// IsEnabled returns whether email sending is enabled
func (m *MockEmailService) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// SendOTP sends an OTP email to the specified recipient
func (m *MockEmailService) SendOTP(ctx context.Context, to, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return &EmailError{Message: "Mock email service forced error"}
	}

	if !m.enabled {
		return nil
	}

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	// Record the OTP
	m.sentOTPs[to] = otp

	// Record the attempt
	attempt := SendAttempt{
		To:      to,
		Subject: "Nyx - Email Verification",
		Body:    fmt.Sprintf("Verification code: %s", otp),
		Time:    time.Now(),
		Type:    "OTP",
		OTP:     otp,
	}
	m.sendAttempts = append(m.sendAttempts, attempt)

	return nil
}

// SendPasswordReset sends a password reset OTP email to the specified recipient
func (m *MockEmailService) SendPasswordReset(ctx context.Context, to, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return &EmailError{Message: "Mock email service forced error"}
	}

	if !m.enabled {
		return nil
	}

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	// Record the reset OTP
	m.sentResets[to] = otp

	// Record the attempt
	attempt := SendAttempt{
		To:      to,
		Subject: "Nyx - Password Reset",
		Body:    fmt.Sprintf("Password reset code: %s", otp),
		Time:    time.Now(),
		Type:    "RESET",
		OTP:     otp,
	}
	m.sendAttempts = append(m.sendAttempts, attempt)

	return nil
}

// GetSentOTP returns the last OTP sent to the specified email
func (m *MockEmailService) GetSentOTP(email string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	otp, exists := m.sentOTPs[email]
	return otp, exists
}

// GetSentPasswordReset returns the last password reset OTP sent to the specified email
func (m *MockEmailService) GetSentPasswordReset(email string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	otp, exists := m.sentResets[email]
	return otp, exists
}

// GetSendAttempts returns all recorded send attempts
func (m *MockEmailService) GetSendAttempts() []SendAttempt {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	attempts := make([]SendAttempt, len(m.sendAttempts))
	copy(attempts, m.sendAttempts)
	return attempts
}

// GetAttemptsByEmail returns send attempts for a specific email
func (m *MockEmailService) GetAttemptsByEmail(email string) []SendAttempt {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var attempts []SendAttempt
	for _, attempt := range m.sendAttempts {
		if attempt.To == email {
			attempts = append(attempts, attempt)
		}
	}
	return attempts
}

// Clear clears all stored data (useful for test cleanup)
func (m *MockEmailService) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentOTPs = make(map[string]string)
	m.sentResets = make(map[string]string)
	m.sendAttempts = make([]SendAttempt, 0)
}

// EmailError is a custom error type for email operations
type EmailError struct {
	Message string
}

func (e *EmailError) Error() string {
	return e.Message
}
