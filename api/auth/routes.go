package api

import (
	"github.com/KiranRajeev-KV/nyx-backend/internal/email"
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

var EmailService email.IEmailService

func InitAuthRoutes(emailService email.IEmailService) {
	EmailService = emailService
}

func AuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", RegisterUser)
		auth.POST("/verify-otp", mw.TempAuth, VerifyOTP)
		auth.POST("/resend-otp", mw.TempAuth, ResendOTP)
		auth.POST("/login", LoginUser)

		auth.POST("/refresh", RefreshToken)
		auth.POST("/forgot-password", ForgotPassword)
		auth.POST("/reset-password", mw.TempAuth, ResetPassword)
		auth.GET("/session", mw.Auth, FetchUserSession)
		auth.GET("/logout", mw.Auth, LogoutUser)

	}
}
