package pkg_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Note: TestMain is defined in token_test.go for the pkg_test package

// TestRequest is a simple validatable request for testing
type TestRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (r TestRequest) Validate() (errorMsg string, err error) {
	if r.Name == "" {
		return "Name is required", assert.AnError
	}
	if r.Email == "" {
		return "Email is required", assert.AnError
	}
	return "", nil
}

// TestRequestAlwaysValid always passes validation
type TestRequestAlwaysValid struct {
	Data string `json:"data"`
}

func (r TestRequestAlwaysValid) Validate() (errorMsg string, err error) {
	return "", nil
}

func createTestContextWithBody(body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

// ==================== ValidateRequest Tests ====================

func TestValidateRequest_ValidJSON_ReturnsRequest(t *testing.T) {
	c, w := createTestContextWithBody(`{"name": "John", "email": "john@example.com"}`)

	req, ok := pkg.ValidateRequest[TestRequest](c)

	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, "John", req.Name)
	assert.Equal(t, "john@example.com", req.Email)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRequest_InvalidJSON_ReturnsFalse(t *testing.T) {
	c, w := createTestContextWithBody(`{invalid json}`)

	req, ok := pkg.ValidateRequest[TestRequest](c)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateRequest_EmptyBody_ReturnsFalse(t *testing.T) {
	c, w := createTestContextWithBody(``)

	req, ok := pkg.ValidateRequest[TestRequest](c)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateRequest_ValidationFails_ReturnsFalse(t *testing.T) {
	// Missing required "name" field
	c, w := createTestContextWithBody(`{"email": "john@example.com"}`)

	req, ok := pkg.ValidateRequest[TestRequest](c)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateRequest_ValidationPasses_ReturnsTrue(t *testing.T) {
	c, w := createTestContextWithBody(`{"data": "anything"}`)

	req, ok := pkg.ValidateRequest[TestRequestAlwaysValid](c)

	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, "anything", req.Data)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRequest_ExtraFields_IgnoredAndPasses(t *testing.T) {
	c, w := createTestContextWithBody(`{"name": "John", "email": "john@example.com", "extra": "ignored"}`)

	req, ok := pkg.ValidateRequest[TestRequest](c)

	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, "John", req.Name)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRequest_WrongContentType_ReturnsFalse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name": "John"}`))
	c.Request.Header.Set("Content-Type", "text/plain")

	req, ok := pkg.ValidateRequest[TestRequest](c)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
