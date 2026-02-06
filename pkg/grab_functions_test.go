package pkg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func newTestContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func TestTagRequestWithId_SetsRequestID(t *testing.T) {
	router := gin.New()
	router.Use(pkg.TagRequestWithId)

	var requestID string
	router.GET("/test", func(c *gin.Context) {
		val, ok := c.Get("request_id")
		assert.True(t, ok)
		requestID, _ = val.(string)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, requestID)
	_, err := ksuid.Parse(requestID)
	assert.NoError(t, err)
}

func TestGetEmail_MissingEmail_ReturnsFalse(t *testing.T) {
	c, w := newTestContext(http.MethodGet, "/test")

	email, ok := pkg.GetEmail(c, "TEST")
	assert.False(t, ok)
	assert.Empty(t, email)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetEmail_HasEmail_ReturnsValue(t *testing.T) {
	c, _ := newTestContext(http.MethodGet, "/test")
	c.Set("email", "test@example.com")

	email, ok := pkg.GetEmail(c, "TEST")
	assert.True(t, ok)
	assert.Equal(t, "test@example.com", email)
}

func TestGrabUserId_MissingUserId_ReturnsFalse(t *testing.T) {
	c, w := newTestContext(http.MethodGet, "/test")

	userID, ok := pkg.GrabUserId(c, "TEST")
	assert.False(t, ok)
	assert.Empty(t, userID)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGrabUserId_HasUserId_ReturnsValue(t *testing.T) {
	c, _ := newTestContext(http.MethodGet, "/test")
	c.Set("userId", "user-123")

	userID, ok := pkg.GrabUserId(c, "TEST")
	assert.True(t, ok)
	assert.Equal(t, "user-123", userID)
}

func TestGrabUuid_InvalidUUID_ReturnsFalse(t *testing.T) {
	c, w := newTestContext(http.MethodGet, "/test")

	parsed, ok := pkg.GrabUuid(c, "not-a-uuid", "TEST", "item")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, parsed)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGrabUuid_ValidUUID_ReturnsTrue(t *testing.T) {
	c, _ := newTestContext(http.MethodGet, "/test")

	id := uuid.New()
	parsed, ok := pkg.GrabUuid(c, id.String(), "TEST", "item")
	assert.True(t, ok)
	assert.Equal(t, id, parsed)
}
