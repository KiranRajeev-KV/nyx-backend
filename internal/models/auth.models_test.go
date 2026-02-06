package models_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// ==================== RegisterUserRequest Tests ====================

func TestRegisterUserRequest_Valid_NoError(t *testing.T) {
	req := models.RegisterUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	_, err := req.Validate()

	assert.NoError(t, err)
}

func TestRegisterUserRequest_MissingName_ReturnsError(t *testing.T) {
	req := models.RegisterUserRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	msg, err := req.Validate()

	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for registering a user", msg)
}

func TestRegisterUserRequest_NameTooShort_ReturnsError(t *testing.T) {
	req := models.RegisterUserRequest{
		Name:     "Jo",
		Email:    "john@example.com",
		Password: "password123",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

func TestRegisterUserRequest_InvalidEmail_ReturnsError(t *testing.T) {
	req := models.RegisterUserRequest{
		Name:     "John Doe",
		Email:    "not-an-email",
		Password: "password123",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

func TestRegisterUserRequest_PasswordTooShort_ReturnsError(t *testing.T) {
	req := models.RegisterUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "short",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

// ==================== LoginUserRequest Tests ====================

func TestLoginUserRequest_Valid_NoError(t *testing.T) {
	req := models.LoginUserRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	_, err := req.Validate()

	assert.NoError(t, err)
}

func TestLoginUserRequest_MissingEmail_ReturnsError(t *testing.T) {
	req := models.LoginUserRequest{
		Password: "password123",
	}

	msg, err := req.Validate()

	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for logging in a user", msg)
}

func TestLoginUserRequest_MissingPassword_ReturnsError(t *testing.T) {
	req := models.LoginUserRequest{
		Email: "john@example.com",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

// ==================== VerifyOTPRequest Tests ====================

func TestVerifyOTPRequest_Valid_NoError(t *testing.T) {
	req := models.VerifyOTPRequest{
		OTP: "123456",
	}

	_, err := req.Validate()

	assert.NoError(t, err)
}

func TestVerifyOTPRequest_MissingOTP_ReturnsError(t *testing.T) {
	req := models.VerifyOTPRequest{}

	msg, err := req.Validate()

	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for verifying OTP", msg)
}

func TestVerifyOTPRequest_OTPTooShort_ReturnsError(t *testing.T) {
	req := models.VerifyOTPRequest{
		OTP: "12345",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

func TestVerifyOTPRequest_OTPTooLong_ReturnsError(t *testing.T) {
	req := models.VerifyOTPRequest{
		OTP: "1234567",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

// ==================== ForgotPasswordRequest Tests ====================

func TestForgotPasswordRequest_Valid_NoError(t *testing.T) {
	req := models.ForgotPasswordRequest{
		Email: "john@example.com",
	}

	_, err := req.Validate()

	assert.NoError(t, err)
}

func TestForgotPasswordRequest_MissingEmail_ReturnsError(t *testing.T) {
	req := models.ForgotPasswordRequest{}

	msg, err := req.Validate()

	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for forgot password", msg)
}

func TestForgotPasswordRequest_InvalidEmail_ReturnsError(t *testing.T) {
	req := models.ForgotPasswordRequest{
		Email: "not-an-email",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

// ==================== ResetPasswordRequest Tests ====================

func TestResetPasswordRequest_Valid_NoError(t *testing.T) {
	req := models.ResetPasswordRequest{
		OTP:      "123456",
		Password: "newpassword123",
	}

	_, err := req.Validate()

	assert.NoError(t, err)
}

func TestResetPasswordRequest_MissingOTP_ReturnsError(t *testing.T) {
	req := models.ResetPasswordRequest{
		Password: "newpassword123",
	}

	msg, err := req.Validate()

	assert.Error(t, err)
	assert.Equal(t, "Invalid request format for reset password", msg)
}

func TestResetPasswordRequest_MissingPassword_ReturnsError(t *testing.T) {
	req := models.ResetPasswordRequest{
		OTP: "123456",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}

func TestResetPasswordRequest_PasswordTooShort_ReturnsError(t *testing.T) {
	req := models.ResetPasswordRequest{
		OTP:      "123456",
		Password: "short",
	}

	_, err := req.Validate()

	assert.Error(t, err)
}
