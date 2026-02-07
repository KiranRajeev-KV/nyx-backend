package tests

import (
	"github.com/gin-gonic/gin"
)

// SetAuthContext sets up an authenticated user context for testing
// This simulates what auth middleware would do
func SetAuthContext(c *gin.Context, email string) {
	c.Set("email", email)
}

// SetTempTokenContext sets up a temporary token context (for OTP flow)
// This simulates what the temp token middleware would do
func SetTempTokenContext(c *gin.Context, email string) {
	c.Set("email", email)
}

// NewTestContextWithAuth creates a test context with authentication already set
func NewTestContextWithAuth(method, path, email string) *TestContext {
	tc := NewTestContext(method, path)
	SetAuthContext(tc.Context, email)
	return tc
}

// NewTestContextWithTempToken creates a test context with temp token already set
func NewTestContextWithTempToken(method, path, email string) *TestContext {
	tc := NewTestContext(method, path)
	SetTempTokenContext(tc.Context, email)
	return tc
}

// NewTestRouterWithAuthMiddleware creates a router with a simple auth middleware
// that expects "X-Test-Email" header and sets it in context
func NewTestRouterWithAuthMiddleware() *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		email := c.GetHeader("X-Test-Email")
		if email != "" {
			SetAuthContext(c, email)
		}
		c.Next()
	})
	return router
}

// NewTestRouterWithTempTokenMiddleware creates a router with a simple temp token middleware
// that expects "X-Test-Temp-Email" header and sets it in context
func NewTestRouterWithTempTokenMiddleware() *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		email := c.GetHeader("X-Test-Temp-Email")
		if email != "" {
			SetTempTokenContext(c, email)
		}
		c.Next()
	})
	return router
}

// ExecuteAuthenticatedRequest executes a request through a router with authentication
// Automatically sets the X-Test-Email header
func ExecuteAuthenticatedRequest(router *gin.Engine, method, path, email string) *TestContext {
	tc := NewTestContext(method, path)
	tc.Context.Request.Header.Set("X-Test-Email", email)
	router.ServeHTTP(tc.Writer, tc.Context.Request)
	return tc
}

// ExecuteAuthenticatedRequestWithJSONBody executes a JSON request with authentication
// Automatically sets the X-Test-Email header
func ExecuteAuthenticatedRequestWithJSONBody(router *gin.Engine, method, path, email string, body interface{}) *TestContext {
	tc := ExecuteRequestWithJSONBody(router, method, path, body)
	tc.Context.Request.Header.Set("X-Test-Email", email)
	w := newTestContext(method, path).Writer // Fresh response recorder
	router.ServeHTTP(w, tc.Context.Request)
	tc.Writer = w
	return tc
}

// ExecuteTempTokenRequest executes a request through a router with temp token
// Automatically sets the X-Test-Temp-Email header
func ExecuteTempTokenRequest(router *gin.Engine, method, path, email string) *TestContext {
	tc := NewTestContext(method, path)
	tc.Context.Request.Header.Set("X-Test-Temp-Email", email)
	router.ServeHTTP(tc.Writer, tc.Context.Request)
	return tc
}

// ExecuteTempTokenRequestWithJSONBody executes a JSON request with temp token
// Automatically sets the X-Test-Temp-Email header
func ExecuteTempTokenRequestWithJSONBody(router *gin.Engine, method, path, email string, body interface{}) *TestContext {
	tc := ExecuteRequestWithJSONBody(router, method, path, body)
	tc.Context.Request.Header.Set("X-Test-Temp-Email", email)
	w := newTestContext(method, path).Writer // Fresh response recorder
	router.ServeHTTP(w, tc.Context.Request)
	tc.Writer = w
	return tc
}

// newTestContext is a private helper to get a fresh context
func newTestContext(method, path string) *TestContext {
	return NewTestContext(method, path)
}
