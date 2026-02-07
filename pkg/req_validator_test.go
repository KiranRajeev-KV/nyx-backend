package pkg_test

import (
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/tests"
	"github.com/stretchr/testify/assert"
)

func init() {
	tests.InitTestLogger()
}

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

// ==================== ValidateRequest Tests ====================

func TestValidateRequest_ValidJSON_ReturnsRequest(t *testing.T) {
	tc := tests.NewTestContextWithBody("POST", "/", `{"name": "John", "email": "john@example.com"}`, "application/json")

	req, ok := pkg.ValidateRequest[TestRequest](tc.Context)

	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, "John", req.Name)
	assert.Equal(t, "john@example.com", req.Email)
	assert.Equal(t, 200, tc.GetResponseStatus())
}

func TestValidateRequest_InvalidJSON_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContextWithBody("POST", "/", `{invalid json}`, "application/json")

	req, ok := pkg.ValidateRequest[TestRequest](tc.Context)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, 400, tc.GetResponseStatus())
}

func TestValidateRequest_EmptyBody_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContextWithBody("POST", "/", ``, "application/json")

	req, ok := pkg.ValidateRequest[TestRequest](tc.Context)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, 400, tc.GetResponseStatus())
}

func TestValidateRequest_ValidationFails_ReturnsFalse(t *testing.T) {
	// Missing required "name" field
	tc := tests.NewTestContextWithBody("POST", "/", `{"email": "john@example.com"}`, "application/json")

	req, ok := pkg.ValidateRequest[TestRequest](tc.Context)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, 400, tc.GetResponseStatus())
}

func TestValidateRequest_ValidationPasses_ReturnsTrue(t *testing.T) {
	tc := tests.NewTestContextWithBody("POST", "/", `{"data": "anything"}`, "application/json")

	req, ok := pkg.ValidateRequest[TestRequestAlwaysValid](tc.Context)

	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, "anything", req.Data)
	assert.Equal(t, 200, tc.GetResponseStatus())
}

func TestValidateRequest_ExtraFields_IgnoredAndPasses(t *testing.T) {
	tc := tests.NewTestContextWithBody("POST", "/", `{"name": "John", "email": "john@example.com", "extra": "ignored"}`, "application/json")

	req, ok := pkg.ValidateRequest[TestRequest](tc.Context)

	assert.True(t, ok)
	assert.NotNil(t, req)
	assert.Equal(t, "John", req.Name)
	assert.Equal(t, 200, tc.GetResponseStatus())
}

func TestValidateRequest_WrongContentType_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContextWithBody("POST", "/", `{"name": "John"}`, "text/plain")

	req, ok := pkg.ValidateRequest[TestRequest](tc.Context)

	assert.False(t, ok)
	assert.Nil(t, req)
	assert.Equal(t, 400, tc.GetResponseStatus())
}
