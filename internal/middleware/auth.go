package mw

import (
	"context"
	"net/http"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func Auth(c *gin.Context) {

	// Extract refresh token
	refreshToken, refErr := c.Cookie("refresh_token")
	if refErr == http.ErrNoCookie {
		pkg.NullifyCookies(c)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Access denied.",
		})
		logger.Log.ErrorCtx(c, "[AUTH-ERROR]: Refresh Cookie is missing", refErr)
		return
	}

	// FLOW: Extract AuthToken
	// 1. If AuthToken available then verify it and set the gin.Context
	// 2. If not available then check against DB and see if a refresh token
	// exists there and if it is a valid one or not
	// 3. If the token is valid then new authToken can be minted, added to
	// the cookie and the gin.Context be populate as well

	accessToken, accessErr := c.Cookie("access_token")
	if accessErr == nil && pkg.VerifyTokens(c, accessToken, refreshToken) {
		// Check if user is banned
		if checkBanned(c) {
			return
		}
		c.Next()
		return
	}

	// If we reach here, either the access token is missing, or it is expired/invalid.
	// In both cases, we should attempt to mint a new one using the refresh token.

	validToken, err := pkg.VerifyRefreshToken(c, refreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Access denied. Session expired.",
		})
		logger.Log.ErrorCtx(c, "[COOKIE-ERROR]: Failed to verify refresh token", err)
		return
	}

	refreshTokenClaims := validToken.Claims()
	userId, _ := refreshTokenClaims["aud"].(string)
	email, _ := refreshTokenClaims["jti"].(string)
	role := db.UserRole(refreshTokenClaims["role"].(string))

	// Creating and setting auth token, so it can be used for future requests
	authToken, err := pkg.CreateAuthToken(userId, email, role)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		logger.Log.FatalCtx(c, "[COOKIE-ERROR]: Failed to mint new auth token", err)
		return
	}
	pkg.SetAuthCookie(c, authToken)

	c.Set("userId", userId)
	c.Set("email", email)
	c.Set("role", string(role))

	// Check if user is banned
	if checkBanned(c) {
		return
	}

	c.Next()
}

// checkBanned looks up the user's ban status from the DB.
// Returns true if the user is banned (and response was already sent).
func checkBanned(c *gin.Context) bool {
	userIdStr, ok := c.Get("userId")
	if !ok {
		return false
	}

	userUUID, err := uuid.Parse(userIdStr.(string))
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		logger.Log.ErrorCtx(c, "[AUTH-BAN-CHECK] Failed to acquire DB connection", err)
		return false // fail open — don't block if DB is unavailable
	}
	defer conn.Release()

	q := db.New()
	isBanned, err := q.CheckUserBanned(ctx, conn, userUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// User no longer exists in database (e.g., after DB reset)
			pkg.NullifyCookies(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Session expired or invalid. Please log in again.",
			})
			logger.Log.WarnCtx(c, "[AUTH-BAN-CHECK] Access attempted with non-existent user ID")
			return true // stop request
		}
		logger.Log.ErrorCtx(c, "[AUTH-BAN-CHECK] Failed to check ban status", err)
		return false
	}

	if isBanned {
		pkg.NullifyCookies(c)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Your account has been suspended. Contact support for assistance.",
		})
		logger.Log.WarnCtx(c, "[AUTH-BAN-CHECK] Banned user attempted access")
		return true
	}

	return false
}

func TempAuth(c *gin.Context) {
	tempToken, err := c.Cookie("temp_token")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Access denied.",
		})
		logger.Log.ErrorCtx(c, "[REQ-ERROR]: Missing temporary auth token", err)
		return
	}

	if !pkg.VerifyTempToken(c, tempToken) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "User is forbidden",
		})
		logger.Log.ErrorCtx(c, "[REQ-ERROR]: Temporary token could not be verified", err)
		return
	}

	c.Next()
}

func CheckUserRole(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		logger.Log.WarnCtx(c, "[ROLE-ERROR]: Could not extract role from context")
		return
	}

	if role == "USER" {
		c.Next()
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"message": "Insufficient permissions to access this resource.",
	})
	logger.Log.WarnCtx(c, "[ROLE-ERROR]: Account does not have USER role")
}

func CheckAdminRole(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		logger.Log.WarnCtx(c, "[ROLE-ERROR]: Could not extract role from context")
		return
	}

	if role == "ADMIN" {
		c.Next()
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"message": "Insufficient permissions to access this resource.",
	})
	logger.Log.WarnCtx(c, "[ROLE-ERROR]: Account does not have ADMIN role")
}
