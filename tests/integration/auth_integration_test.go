package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to extract a cookie from response
func getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// Retrieves OTP from DB for testing verification
func getLatestOTP(t *testing.T, email string) string {
	var otp string
	err := testDBPool.QueryRow(context.Background(), "SELECT otp FROM user_onboarding WHERE email = $1 ORDER BY created_at DESC LIMIT 1", email).Scan(&otp)
	require.NoError(t, err, "Failed to fetch OTP from DB")
	return otp
}

func TestAuthFlow_Integration(t *testing.T) {
	cleanDB(t)

	// User details
	testUser := models.RegisterUserRequest{
		Name:     "Integration User",
		Email:    "int_user@gmail.com",
		Password: "SecurePassword123!",
	}

	var tempAuthCookie *http.Cookie
	var accessCookie *http.Cookie
	var refreshCookie *http.Cookie

	t.Run("INT-USER-001: Register User", func(t *testing.T) {
		body, _ := json.Marshal(testUser)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify temp cookie was set
		cookies := w.Result().Cookies()
		tempAuthCookie = getCookie(cookies, "temp_token")
		require.NotNil(t, tempAuthCookie, "Temp auth cookie should be set after register")
	})

	t.Run("INT-USER-002: Verify OTP", func(t *testing.T) {
		// Fetch the actual OTP created in the database
		otp := getLatestOTP(t, testUser.Email)

		verifyReq := models.VerifyOTPRequest{
			OTP: otp,
		}
		body, _ := json.Marshal(verifyReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/verify-otp", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(tempAuthCookie)

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify temp cookie is cleared (MaxAge < 0 or empty value)
		cookies := w.Result().Cookies()
		clearedCookie := getCookie(cookies, "temp_token")
		if clearedCookie != nil {
			assert.Equal(t, "", clearedCookie.Value, "Temp auth cookie should be cleared")
		}
	})

	t.Run("INT-USER-003: Login", func(t *testing.T) {
		loginReq := models.LoginUserRequest{
			Email:    testUser.Email,
			Password: testUser.Password,
		}
		body, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		accessCookie = getCookie(cookies, "access_token")
		refreshCookie = getCookie(cookies, "refresh_token")

		require.NotNil(t, accessCookie, "Access cookie must be set")
		require.NotNil(t, refreshCookie, "Refresh cookie must be set")
	})

	t.Run("INT-USER-004: Protected Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Session route requires auth middleware
		req, _ := http.NewRequest("GET", "/api/v1/auth/session", nil)
		req.AddCookie(accessCookie)
		req.AddCookie(refreshCookie)

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assuming standard response format: {"success": true, "data": {"email": "..."}}
		assert.Equal(t, testUser.Email, response["email"])
	})

	t.Run("INT-USER-005: Refresh Flow", func(t *testing.T) {
		// Wait a second to ensure token issue time is slightly different if needed
		time.Sleep(1 * time.Second)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		// Provide ONLY the refresh cookie
		req.AddCookie(refreshCookie)

		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		newAccessCookie := getCookie(cookies, "access_token")
		require.NotNil(t, newAccessCookie, "New access cookie must be issued")

		// Ensure the new access token is different from the old one
		assert.NotEqual(t, accessCookie.Value, newAccessCookie.Value)

		// Test protected route with new access token
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/v1/auth/session", nil)
		req2.AddCookie(newAccessCookie)
		req2.AddCookie(refreshCookie)
		testRouter.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code, "Should be authorized with new token")
	})
}
