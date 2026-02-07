package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// TestContext wraps a Gin context with its response recorder for testing
type TestContext struct {
	Context *gin.Context
	Writer  *httptest.ResponseRecorder
}

// NewTestContext creates a new test context with the given HTTP method and path
func NewTestContext(method, path string) *TestContext {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return &TestContext{
		Context: c,
		Writer:  w,
	}
}

// NewTestContextWithBody creates a test context with a request body and content-type
func NewTestContextWithBody(method, path string, body string, contentType ...string) *TestContext {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)

	if body != "" {
		c.Request.Body = io.NopCloser(bytes.NewBufferString(body))
	}

	if len(contentType) > 0 {
		c.Request.Header.Set("Content-Type", contentType[0])
	}

	return &TestContext{
		Context: c,
		Writer:  w,
	}
}

// NewTestRouter creates a new Gin router for testing
func NewTestRouter() *gin.Engine {
	return gin.New()
}

// NewTestRouterWithMiddleware creates a Gin router with middleware for testing
func NewTestRouterWithMiddleware(middleware ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	for _, m := range middleware {
		router.Use(m)
	}
	return router
}

// SetContextValue is a helper to set values in the test context
func (tc *TestContext) SetContextValue(key string, value interface{}) {
	tc.Context.Set(key, value)
}

// GetContextValue is a helper to get values from the test context
func (tc *TestContext) GetContextValue(key string) (interface{}, bool) {
	return tc.Context.Get(key)
}

// GetResponseStatus returns the HTTP status code from the response
func (tc *TestContext) GetResponseStatus() int {
	return tc.Writer.Code
}

// GetResponseBody returns the response body as a string
func (tc *TestContext) GetResponseBody() string {
	return tc.Writer.Body.String()
}

// ExecuteRequest executes a request through a router and returns a TestContext
func ExecuteRequest(router *gin.Engine, method, path string) *TestContext {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	router.ServeHTTP(w, req)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return &TestContext{
		Context: c,
		Writer:  w,
	}
}

// ExecuteRequestWithJSONBody executes a request with JSON body through a router
func ExecuteRequestWithJSONBody(router *gin.Engine, method, path string, body interface{}) *TestContext {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return &TestContext{
		Context: c,
		Writer:  w,
	}
}
