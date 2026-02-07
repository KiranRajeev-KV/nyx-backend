package pkg_test

import (
	"os"
	"testing"

	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/tests"
	"github.com/stretchr/testify/assert"
)

func init() {
	tests.InitTestLogger()
	
	// Change to project root to load RSA keys
	os.Chdir("/home/kr/dev/nyx-backend")
	pkg.InitPaseto()
}

func TestParseToken_InvalidTokenType_ReturnsFalse(t *testing.T) {
	valid, parsedToken := pkg.ParseToken("invalid.token.here", "invalid_type")

	assert.False(t, valid)
	assert.Nil(t, parsedToken)
}

func TestParseToken_EmptyToken_ReturnsFalse(t *testing.T) {
	valid, parsedToken := pkg.ParseToken("", "access_token")

	assert.False(t, valid)
	assert.Nil(t, parsedToken)
}

func TestParseToken_MalformedToken_ReturnsFalse(t *testing.T) {
	valid, parsedToken := pkg.ParseToken("this.is.not.a.valid.paseto.token", "access_token")

	assert.False(t, valid)
	assert.Nil(t, parsedToken)
}

func TestParseToken_ValidAccessToken_ReturnsTrue(t *testing.T) {
	token, err := pkg.CreateAuthToken("user-123", "test@example.com", db.UserRoleUSER)
	assert.NoError(t, err)

	valid, parsedToken := pkg.ParseToken(token, "access_token")

	assert.True(t, valid)
	assert.NotNil(t, parsedToken)
}

func TestParseToken_ValidRefreshToken_ReturnsTrue(t *testing.T) {
	token, err := pkg.CreateRefreshToken("user-123", "test@example.com", db.UserRoleUSER)
	assert.NoError(t, err)

	valid, parsedToken := pkg.ParseToken(token, "refresh_token")

	assert.True(t, valid)
	assert.NotNil(t, parsedToken)
}

func TestParseToken_WrongTokenType_ReturnsFalse(t *testing.T) {
	// Create an access token but try to parse it as refresh token
	token, err := pkg.CreateAuthToken("user-123", "test@example.com", db.UserRoleUSER)
	assert.NoError(t, err)

	valid, parsedToken := pkg.ParseToken(token, "refresh_token")

	assert.False(t, valid)
	assert.Nil(t, parsedToken)
}

func TestVerifyTokens_EmptyTokens_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context

	result := pkg.VerifyTokens(c, "", "")

	assert.False(t, result)
}

func TestVerifyTokens_EmptyAuthToken_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	refreshToken, _ := pkg.CreateRefreshToken("user-123", "test@example.com", db.UserRoleUSER)

	result := pkg.VerifyTokens(c, "", refreshToken)

	assert.False(t, result)
}

func TestVerifyTokens_EmptyRefreshToken_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	authToken, _ := pkg.CreateAuthToken("user-123", "test@example.com", db.UserRoleUSER)

	result := pkg.VerifyTokens(c, authToken, "")

	assert.False(t, result)
}

func TestVerifyTokens_ValidTokens_ReturnsTrue(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	userId := "user-123"
	email := "test@example.com"
	role := db.UserRoleUSER

	authToken, err := pkg.CreateAuthToken(userId, email, role)
	assert.NoError(t, err)
	refreshToken, err := pkg.CreateRefreshToken(userId, email, role)
	assert.NoError(t, err)

	result := pkg.VerifyTokens(c, authToken, refreshToken)

	assert.True(t, result)
}

func TestVerifyTokens_SetsContextValues(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	userId := "user-123"
	email := "test@example.com"
	role := db.UserRoleUSER

	authToken, _ := pkg.CreateAuthToken(userId, email, role)
	refreshToken, _ := pkg.CreateRefreshToken(userId, email, role)

	pkg.VerifyTokens(c, authToken, refreshToken)

	ctxUserId, exists := c.Get("userId")
	assert.True(t, exists)
	assert.Equal(t, userId, ctxUserId)

	ctxEmail, exists := c.Get("email")
	assert.True(t, exists)
	assert.Equal(t, email, ctxEmail)

	ctxRole, exists := c.Get("role")
	assert.True(t, exists)
	assert.Equal(t, string(role), ctxRole)
}

func TestVerifyTokens_MismatchedUserId_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	email := "test@example.com"
	role := db.UserRoleUSER

	authToken, _ := pkg.CreateAuthToken("user-123", email, role)
	refreshToken, _ := pkg.CreateRefreshToken("user-456", email, role) // Different userId

	result := pkg.VerifyTokens(c, authToken, refreshToken)

	assert.False(t, result)
}

func TestVerifyTokens_MismatchedEmail_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	userId := "user-123"
	role := db.UserRoleUSER

	authToken, _ := pkg.CreateAuthToken(userId, "test1@example.com", role)
	refreshToken, _ := pkg.CreateRefreshToken(userId, "test2@example.com", role) // Different email

	result := pkg.VerifyTokens(c, authToken, refreshToken)

	assert.False(t, result)
}

func TestVerifyTokens_MismatchedRole_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	userId := "user-123"
	email := "test@example.com"

	authToken, _ := pkg.CreateAuthToken(userId, email, db.UserRoleUSER)
	refreshToken, _ := pkg.CreateRefreshToken(userId, email, db.UserRoleADMIN) // Different role

	result := pkg.VerifyTokens(c, authToken, refreshToken)

	assert.False(t, result)
}

func TestVerifyTempToken_EmptyToken_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context

	result := pkg.VerifyTempToken(c, "")

	assert.False(t, result)
}

func TestVerifyTempToken_InvalidToken_ReturnsFalse(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context

	result := pkg.VerifyTempToken(c, "invalid.token.here")

	assert.False(t, result)
}

func TestVerifyTempToken_ValidToken_ReturnsTrue(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	email := "test@example.com"

	tempToken := pkg.CreateTempToken(email)

	result := pkg.VerifyTempToken(c, tempToken)

	assert.True(t, result)
}

func TestVerifyTempToken_SetsEmailInContext(t *testing.T) {
	tc := tests.NewTestContext("POST", "/")
	c := tc.Context
	email := "test@example.com"

	tempToken := pkg.CreateTempToken(email)
	pkg.VerifyTempToken(c, tempToken)

	ctxEmail, exists := c.Get("email")
	assert.True(t, exists)
	assert.Equal(t, email, ctxEmail)
}

func TestCreateAuthToken_ReturnsValidToken(t *testing.T) {
	token, err := pkg.CreateAuthToken("user-123", "test@example.com", db.UserRoleUSER)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCreateRefreshToken_ReturnsValidToken(t *testing.T) {
	token, err := pkg.CreateRefreshToken("user-123", "test@example.com", db.UserRoleUSER)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCreateTempToken_ReturnsValidToken(t *testing.T) {
	token := pkg.CreateTempToken("test@example.com")

	assert.NotEmpty(t, token)
}

func TestCreateAuthToken_AdminRole_ReturnsValidToken(t *testing.T) {
	token, err := pkg.CreateAuthToken("admin-123", "admin@example.com", db.UserRoleADMIN)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	valid, _ := pkg.ParseToken(token, "access_token")
	assert.True(t, valid)
}
