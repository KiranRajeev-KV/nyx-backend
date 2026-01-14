package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", RegisterUser)
		auth.POST("/verify-otp", mw.TempAuth, VerifyOTP) // TODO: Update logic of this endpoint
		// TODO: auth.GET("/resend-otp", mw.TempAuth, SendOTP)
		auth.POST("/login", LoginUser)

		// TODO: auth.POST("/refresh", mw.Auth, RefreshToken)
		// TODO: auth.POST("/reset-password", mw.TempAuth, ResetPassword)
		auth.GET("/session", mw.Auth, FetchUserSession)
		auth.GET("/logout", mw.Auth, LogoutUser)

	}
}
