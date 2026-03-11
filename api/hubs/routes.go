package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func HubRoutes(router *gin.RouterGroup) {
	hubs := router.Group("/hubs")
	{
		// Public endpoints
		hubs.GET("", FetchHubs)
		hubs.GET("/:id", FetchHubById)

		// Admin-only endpoints
		hubs.POST("/", mw.Auth, mw.CheckAdminRole, CreateHub)
		hubs.PATCH("/:id", mw.Auth, mw.CheckAdminRole, UpdateHub)
		hubs.DELETE("/:id", mw.Auth, mw.CheckAdminRole, DeleteHub)
	}
}
