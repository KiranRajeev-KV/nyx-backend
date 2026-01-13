package mw

import (
	"net/http"
	"slices"

	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
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
		c.Next()
		return
	}

	if accessErr == http.ErrNoCookie {
		// Check if refresh token is valid. If yes, only then check for access token
		// validity. If access token is valid then setup gin.Context map otherwise
		// mint new token and then setup gin.Context map
		validToken, err := pkg.VerifyRefreshToken(c, refreshToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Access denied.",
			})
			logger.Log.ErrorCtx(c, "[COOKIE-ERROR]: Failed to verify refresh token", err)
			return
		}

		refreshTokenClaims := validToken.Claims()
		userId, _ := refreshTokenClaims["aud"].(string)
		email, _ := refreshTokenClaims["jti"].(string)
		roles := pkg.ParseRolesFromClaims(refreshTokenClaims)

		// Creating and setting auth token, so it can be used for future requests
		authToken, err := pkg.CreateAuthToken(userId, email, roles)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Oops! Something happened. Please try again later.",
			})
			logger.Log.FatalCtx(c, "[COOKIE-ERROR]: Failed to mint new auth token", err)
			return
		}
		pkg.SetAuthCookie(c, authToken)
	}

	c.Next()
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
	roles := c.GetStringSlice("roles")
	if len(roles) == 0 {
		logger.Log.WarnCtx(c, "[ROLE-ERROR]: Could not extract roles from context")
		return
	}
	// if it contains USER role, then allow
	if slices.Contains(roles, "USER") {
		c.Next()
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"message": "Insufficient permissions to access this resource.",
	})
	logger.Log.WarnCtx(c, "[ROLE-ERROR]: Account does not have USER role")
}

func CheckAdminRole(c *gin.Context) {
	roles := c.GetStringSlice("roles")
	if len(roles) == 0 {
		logger.Log.WarnCtx(c, "[ROLE-ERROR]: Could not extract roles from context")
		return
	}
	// if it contains ADMIN role, then allow
	if slices.Contains(roles, "ADMIN") {
		c.Next()
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"message": "Insufficient permissions to access this resource.",
	})
	logger.Log.WarnCtx(c, "[ROLE-ERROR]: Account does not have ADMIN role")
}
