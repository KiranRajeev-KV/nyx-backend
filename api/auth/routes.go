package api

import "github.com/gin-gonic/gin"

func AuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", RegisterUser)
	}
}
