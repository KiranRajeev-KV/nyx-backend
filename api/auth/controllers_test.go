package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	api "github.com/KiranRajeev-KV/nyx-backend/api/auth"
	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	tests.InitTestLogger()
	// Change to project root for RSA keys
	os.Chdir("/home/kr/dev/nyx-backend")
	pkg.InitPaseto()
}

// setupRegisterTest sets up common test dependencies for register tests
func setupRegisterTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.POST("/auth/register", api.RegisterUser)
	return router
}

func TestRegisterUser_ValidRequest_SuccessfulRegistration(t *testing.T) {
	// Note: This would require full DB setup to test
	// Skipping DB integration in unit tests
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestRegisterUser_InvalidEmail_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	reqBody := models.RegisterUserRequest{
		Name:     "John Doe",
		Email:    "not-an-email",
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_InvalidPassword_TooShort_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	reqBody := models.RegisterUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "short",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_InvalidPassword_TooLong_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	longPassword := "a"
	for i := 0; i < 130; i++ {
		longPassword += "a"
	}
	reqBody := models.RegisterUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: longPassword,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_ShortName_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	reqBody := models.RegisterUserRequest{
		Name:     "Jo",
		Email:    "john@example.com",
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_LongName_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	longName := "a"
	for i := 0; i < 102; i++ {
		longName += "a"
	}
	reqBody := models.RegisterUserRequest{
		Name:     longName,
		Email:    "john@example.com",
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_MissingEmail_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	reqBody := models.RegisterUserRequest{
		Name:     "John Doe",
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_MissingPassword_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	reqBody := models.RegisterUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_MissingName_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	reqBody := models.RegisterUserRequest{
		Email:    "john@example.com",
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/register", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestRegisterUser_InvalidJSON_BadRequest(t *testing.T) {
	router := setupRegisterTest()

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// VerifyOTP Tests

func setupVerifyOTPTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.POST("/auth/verify-otp", api.VerifyOTP)
	return router
}

func TestVerifyOTP_ValidRequest_RequiresDBSetup(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// Skipping DB integration in unit tests
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestVerifyOTP_MissingOTP_BadRequest(t *testing.T) {
	router := setupVerifyOTPTest()

	reqBody := models.VerifyOTPRequest{}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/verify-otp", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestVerifyOTP_OTPTooShort_BadRequest(t *testing.T) {
	router := setupVerifyOTPTest()

	reqBody := models.VerifyOTPRequest{
		OTP: "12345", // 5 digits instead of 6
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/verify-otp", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestVerifyOTP_OTPTooLong_BadRequest(t *testing.T) {
	router := setupVerifyOTPTest()

	reqBody := models.VerifyOTPRequest{
		OTP: "1234567", // 7 digits instead of 6
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/verify-otp", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestVerifyOTP_EmptyOTP_BadRequest(t *testing.T) {
	router := setupVerifyOTPTest()

	reqBody := models.VerifyOTPRequest{
		OTP: "",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/verify-otp", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestVerifyOTP_InvalidJSON_BadRequest(t *testing.T) {
	router := setupVerifyOTPTest()

	req := httptest.NewRequest(http.MethodPost, "/auth/verify-otp", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// LoginUser Tests

func setupLoginTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.POST("/auth/login", api.LoginUser)
	return router
}

func TestLoginUser_ValidRequest_RequiresDBSetup(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// Skipping DB integration in unit tests
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestLoginUser_InvalidEmail_BadRequest(t *testing.T) {
	router := setupLoginTest()

	reqBody := models.LoginUserRequest{
		Email:    "not-an-email",
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/login", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestLoginUser_PasswordTooShort_BadRequest(t *testing.T) {
	router := setupLoginTest()

	reqBody := models.LoginUserRequest{
		Email:    "john@example.com",
		Password: "short",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/login", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestLoginUser_PasswordTooLong_BadRequest(t *testing.T) {
	router := setupLoginTest()

	longPassword := "a"
	for i := 0; i < 130; i++ {
		longPassword += "a"
	}
	reqBody := models.LoginUserRequest{
		Email:    "john@example.com",
		Password: longPassword,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/login", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestLoginUser_MissingEmail_BadRequest(t *testing.T) {
	router := setupLoginTest()

	reqBody := models.LoginUserRequest{
		Password: "SecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/login", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestLoginUser_MissingPassword_BadRequest(t *testing.T) {
	router := setupLoginTest()

	reqBody := models.LoginUserRequest{
		Email: "john@example.com",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/login", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestLoginUser_InvalidJSON_BadRequest(t *testing.T) {
	router := setupLoginTest()

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ForgotPassword Tests

func setupForgotPasswordTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.POST("/auth/forgot-password", api.ForgotPassword)
	return router
}

func TestForgotPassword_ValidRequest_RequiresDBSetup(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// Skipping DB integration in unit tests
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestForgotPassword_InvalidEmail_BadRequest(t *testing.T) {
	router := setupForgotPasswordTest()

	reqBody := models.ForgotPasswordRequest{
		Email: "not-an-email",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/forgot-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestForgotPassword_MissingEmail_BadRequest(t *testing.T) {
	router := setupForgotPasswordTest()

	reqBody := models.ForgotPasswordRequest{}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/forgot-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestForgotPassword_EmptyEmail_BadRequest(t *testing.T) {
	router := setupForgotPasswordTest()

	reqBody := models.ForgotPasswordRequest{
		Email: "",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/forgot-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestForgotPassword_InvalidJSON_BadRequest(t *testing.T) {
	router := setupForgotPasswordTest()

	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ResetPassword Tests

func setupResetPasswordTest() *gin.Engine {
	router := tests.NewTestRouter()
	router.POST("/auth/reset-password", api.ResetPassword)
	return router
}

func TestResetPassword_ValidRequest_RequiresDBSetup(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// Skipping DB integration in unit tests
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

func TestResetPassword_InvalidOTP_TooShort_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	reqBody := models.ResetPasswordRequest{
		OTP:      "12345", // 5 digits instead of 6
		Password: "NewSecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/reset-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestResetPassword_InvalidOTP_TooLong_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	reqBody := models.ResetPasswordRequest{
		OTP:      "1234567", // 7 digits instead of 6
		Password: "NewSecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/reset-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestResetPassword_MissingOTP_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	reqBody := models.ResetPasswordRequest{
		Password: "NewSecurePassword123",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/reset-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestResetPassword_InvalidPassword_TooShort_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	reqBody := models.ResetPasswordRequest{
		OTP:      "123456",
		Password: "short",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/reset-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestResetPassword_InvalidPassword_TooLong_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	longPassword := "a"
	for i := 0; i < 130; i++ {
		longPassword += "a"
	}
	reqBody := models.ResetPasswordRequest{
		OTP:      "123456",
		Password: longPassword,
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/reset-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestResetPassword_MissingPassword_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	reqBody := models.ResetPasswordRequest{
		OTP: "123456",
	}

	tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/auth/reset-password", reqBody)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestResetPassword_InvalidJSON_BadRequest(t *testing.T) {
	router := setupResetPasswordTest()

	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// LogoutUser Tests

func setupLogoutTest() *gin.Engine {
	// Router with auth middleware that extracts email from X-Test-Email header
	router := tests.NewTestRouterWithAuthMiddleware()
	router.POST("/auth/logout", api.LogoutUser)
	return router
}

func TestLogoutUser_MissingAuthContext_Fatal(t *testing.T) {
	router := setupLogoutTest()

	// Request without X-Test-Email header (no auth)
	// This will trigger a fatal error in GetEmail since no email is in context
	tc := tests.ExecuteAuthenticatedRequest(router, "POST", "/auth/logout", "")
	// GetEmail logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

func TestLogoutUser_WithValidAuthContext_RequiresDB(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// The middleware context is properly set, but DB operations will fail
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

// ResendOTP Tests

func setupResendOTPTest() *gin.Engine {
	// Router with temp token middleware that extracts email from X-Test-Temp-Email header
	router := tests.NewTestRouterWithTempTokenMiddleware()
	router.POST("/auth/resend-otp", api.ResendOTP)
	return router
}

func TestResendOTP_MissingTempTokenContext_Fatal(t *testing.T) {
	router := setupResendOTPTest()

	// Request without X-Test-Temp-Email header (no temp token)
	// This will trigger a fatal error in GetEmail since no email is in context
	tc := tests.ExecuteTempTokenRequest(router, "POST", "/auth/resend-otp", "")
	// GetEmail logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

func TestResendOTP_ValidTempToken_RequiresDB(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// The middleware context is properly set, but DB operations will fail
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}

// FetchUserSession Tests

func setupFetchUserSessionTest() *gin.Engine {
	// Router with auth middleware that extracts email from X-Test-Email header
	router := tests.NewTestRouterWithAuthMiddleware()
	router.GET("/auth/session", api.FetchUserSession)
	return router
}

func TestFetchUserSession_MissingAuthContext_Fatal(t *testing.T) {
	router := setupFetchUserSessionTest()

	// Request without X-Test-Email header (no auth)
	// This will trigger a fatal error in GetEmail since no email is in context
	tc := tests.ExecuteAuthenticatedRequest(router, "GET", "/auth/session", "")
	// GetEmail logs fatal and returns 500
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

func TestFetchUserSession_ValidAuthContext_RequiresDB(t *testing.T) {
	// Note: This would require full DB setup to test the happy path
	// The middleware context is properly set, but DB operations will fail
	// Full integration tests should use real or test DB
	t.Skip("Requires DB setup - use integration tests instead")
}
