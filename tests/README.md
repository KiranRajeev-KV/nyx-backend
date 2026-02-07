# Test Helpers

This package provides utilities for testing in the Nyx backend project.

## Test Structure

The project follows a structured testing approach:

- **Unit Tests**: Validate request validation and error handling without database
- **Controller Tests**: Located in `api/{domain}/controllers_test.go`
- **Validation Tests**: Test request models and validation rules
- **Integration Tests**: Full end-to-end tests with database

## Gin Testing Helpers

Use `gin.go` for standardized Gin context and router testing:

```go
import "github.com/KiranRajeev-KV/nyx-backend/tests"

// Create a simple test context
tc := tests.NewTestContext(http.MethodGet, "/test")

// Set and get values
tc.SetContextValue("email", "test@example.com")
email, ok := tc.GetContextValue("email")

// Check response
status := tc.GetResponseStatus()
body := tc.GetResponseBody()

// Create a test router with middleware
router := tests.NewTestRouterWithMiddleware(middleware1, middleware2)

// Execute requests through router
tc := tests.ExecuteRequest(router, http.MethodGet, "/test")

// Execute JSON requests
tc := tests.ExecuteRequestWithJSONBody(router, http.MethodPost, "/test", requestBody)
```

## Middleware Testing Helpers

Use `middleware.go` for testing endpoints that require authentication or middleware:

```go
import "github.com/KiranRajeev-KV/nyx-backend/tests"

// Create a router with built-in auth middleware
router := tests.NewTestRouterWithAuthMiddleware()

// Execute request with authentication
tc := tests.ExecuteAuthenticatedRequest(router, http.MethodGet, "/protected", "user@example.com")

// Execute JSON request with authentication
tc := tests.ExecuteAuthenticatedRequestWithJSONBody(router, http.MethodPost, "/protected", "user@example.com", requestBody)

// For temp token flows (OTP, password reset)
router := tests.NewTestRouterWithTempTokenMiddleware()
tc := tests.ExecuteTempTokenRequest(router, http.MethodPost, "/verify-otp", "user@example.com")
tc := tests.ExecuteTempTokenRequestWithJSONBody(router, http.MethodPost, "/verify-otp", "user@example.com", requestBody)

// Or set context directly
tc := tests.NewTestContextWithAuth(http.MethodGet, "/test", "user@example.com")
tc := tests.NewTestContextWithTempToken(http.MethodPost, "/test", "user@example.com")
```

## Logger Testing Helper

Initialize the logger for tests using `logger.go`:

```go
func init() {
    tests.InitTestLogger()  // Discard all logs
    // or
    tests.InitTestLoggerWithOutput(os.Stdout)  // Debug with output
}
```

## Controller Testing Pattern

Each controller has comprehensive unit tests. Tests are organized by endpoint and follow this pattern:

### Setup Functions
Each endpoint group has a setup function that creates a router with necessary routes and middleware:

```go
func setupCreateClaimTest() *gin.Engine {
    router := tests.NewTestRouterWithAuthMiddleware()
    router.POST("/claims", api.CreateClaim)
    return router
}
```

### Test Categories

1. **Validation Tests**: Test request field validation
   ```go
   func TestCreateClaim_InvalidItemID_BadRequest(t *testing.T) {
       router := setupCreateClaimTest()
       reqBody := models.CreateClaimRequest{
           ItemID:    "not-a-uuid",  // Invalid UUID format
           ProofText: "Valid proof text that is long enough",
       }
       tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/claims", reqBody)
       assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
   }
   ```

2. **Length Validation Tests**: Test field length constraints
   ```go
   func TestCreateClaim_ProofTextTooShort_BadRequest(t *testing.T) {
       // Test minimum length requirement
   }
   
   func TestCreateClaim_ProofTextTooLong_BadRequest(t *testing.T) {
       // Test maximum length requirement
   }
   ```

3. **Required Field Tests**: Test missing required fields
   ```go
   func TestCreateClaim_MissingItemID_BadRequest(t *testing.T) {
       // Test when ItemID is not provided
   }
   ```

4. **JSON Parsing Tests**: Test malformed JSON handling
   ```go
   func TestCreateClaim_InvalidJSON_BadRequest(t *testing.T) {
       req := httptest.NewRequest(http.MethodPost, "/claims", 
           bytes.NewBufferString(`{invalid json}`))
       // Test JSON parsing error
   }
   ```

5. **Authentication Tests**: Test auth context requirements
   ```go
   func TestCreateClaim_MissingAuthContext_Fatal(t *testing.T) {
       // Test without authentication context
   }
   ```

6. **Database Setup Tests**: Mark tests requiring DB as skipped
   ```go
   func TestCreateClaim_ValidRequest_RequiresDBSetup(t *testing.T) {
       t.Skip("Requires DB setup - use integration tests instead")
   }
   ```

## Available Controller Tests

### Auth Controller (`api/auth/controllers_test.go`)
- RegisterUser validation tests
- VerifyOTP validation tests
- LoginUser validation tests
- ForgotPassword validation tests
- ResetPassword validation tests
- LogoutUser auth tests
- ResendOTP temp token tests
- FetchUserSession auth tests

### Claims Controller (`api/claims/controllers_test.go`)
- CreateClaim validation tests (ItemID, ProofText, ProofImageUrl)
- FetchUserClaims auth tests
- FetchClaimsByItem validation tests
- FetchAllClaims tests
- ProcessClaim validation tests (Status, AdminNotes)

## Test Statistics

Run all tests:
```bash
go test ./...
```

Current test count: **200+ unit tests** covering:
- Request validation
- Field length constraints
- Required field validation
- JSON parsing
- Authentication context checks
- Error handling

## Usage Pattern

```go
package mypackage_test

import (
    "testing"
    "github.com/KiranRajeev-KV/nyx-backend/tests"
)

func init() {
    tests.InitTestLogger()
    os.Chdir("/home/kr/dev/nyx-backend")
    pkg.InitPaseto()
}

func TestMyFunction(t *testing.T) {
    router := tests.NewTestRouterWithAuthMiddleware()
    router.POST("/my-endpoint", api.MyHandler)
    
    reqBody := models.MyRequest{
        Field: "value",
    }
    
    tc := tests.ExecuteRequestWithJSONBody(router, "POST", "/my-endpoint", reqBody)
    assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}
```

## Best Practices

1. **Skip DB-dependent tests**: Use `t.Skip()` for tests that require database setup
2. **Test validation first**: Prioritize testing request validation in unit tests
3. **Use setup functions**: Create router setup functions for endpoint groups
4. **Mock authentication**: Use test middleware to mock auth without database
5. **Test error paths**: Include tests for all validation failure scenarios
6. **Use assertions**: Leverage testify/assert for clear test output

