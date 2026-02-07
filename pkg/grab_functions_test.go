package pkg_test

import (
	"net/http"
	"testing"

	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/tests"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	tests.InitTestLogger()
}

func TestTagRequestWithId_SetsRequestID(t *testing.T) {
	router := tests.NewTestRouter()
	router.Use(pkg.TagRequestWithId)

	var requestID string
	router.GET("/test", func(c *gin.Context) {
		val, ok := c.Get("request_id")
		assert.True(t, ok)
		requestID, _ = val.(string)
		c.Status(http.StatusOK)
	})

	tc := tests.ExecuteRequest(router, http.MethodGet, "/test")

	assert.Equal(t, http.StatusOK, tc.GetResponseStatus())
	assert.NotEmpty(t, requestID)
	_, err := ksuid.Parse(requestID)
	assert.NoError(t, err)
}

func TestGetEmail_MissingEmail_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext(http.MethodGet, "/test")
	email, ok := pkg.GetEmail(tc.Context, "TEST")
	assert.False(t, ok)
	assert.Empty(t, email)
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

func TestGetEmail_HasEmail_ReturnsValue(t *testing.T) {
	tc := tests.NewTestContext(http.MethodGet, "/test")
	tc.SetContextValue("email", "test@example.com")

	email, ok := pkg.GetEmail(tc.Context, "TEST")
	assert.True(t, ok)
	assert.Equal(t, "test@example.com", email)
}

func TestGrabUserId_MissingUserId_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext(http.MethodGet, "/test")
	userID, ok := pkg.GrabUserId(tc.Context, "TEST")
	assert.False(t, ok)
	assert.Empty(t, userID)
	assert.Equal(t, http.StatusInternalServerError, tc.GetResponseStatus())
}

func TestGrabUserId_HasUserId_ReturnsValue(t *testing.T) {
	tc := tests.NewTestContext(http.MethodGet, "/test")
	tc.SetContextValue("userId", "user-123")

	userID, ok := pkg.GrabUserId(tc.Context, "TEST")
	assert.True(t, ok)
	assert.Equal(t, "user-123", userID)
}

func TestGrabUuid_InvalidUUID_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext(http.MethodGet, "/test")
	parsed, ok := pkg.GrabUuid(tc.Context, "not-a-uuid", "TEST", "item")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, parsed)
	assert.Equal(t, http.StatusBadRequest, tc.GetResponseStatus())
}

func TestGrabUuid_ValidUUID_ReturnsTrue(t *testing.T) {
	tc := tests.NewTestContext(http.MethodGet, "/test")

	id := uuid.New()
	parsed, ok := pkg.GrabUuid(tc.Context, id.String(), "TEST", "item")
	assert.True(t, ok)
	assert.Equal(t, id, parsed)
}
