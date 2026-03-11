package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.RouterGroup) {
	admin := router.Group("/admin")
	admin.Use(mw.Auth, mw.CheckAdminRole)
	{
		admin.GET("/users", FetchAllUsers)
		admin.PATCH("/users/:id/ban", BanUser)
		admin.PATCH("/users/:id/unban", UnbanUser)
		admin.PATCH("/users/:id/promote", PromoteToAdmin)
	}
}
