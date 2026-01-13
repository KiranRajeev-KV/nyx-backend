package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", RegisterUser)
		auth.POST("/verify-otp", mw.TempAuth, VerifyOTP)
		auth.POST("/login", LoginUser)

		auth.GET("/session", mw.Auth, FetchUserSession)
		auth.GET("/logout", mw.Auth, LogoutUser)

	}
}
