package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func ItemRoutes(router *gin.RouterGroup) {
	items := router.Group("/items")
	{
		items.GET("/", mw.Auth, FetchItems)
		items.POST("/", mw.Auth, CreateItem)
		items.GET("/:id", mw.Auth, FetchItemByID)
		items.PUT("/:id", mw.Auth, UpdateItemByID)
		items.DELETE("/:id", mw.Auth, DeleteItemByID)
	}
}
