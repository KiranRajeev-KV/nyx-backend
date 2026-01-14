package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func ItemRoutes(router *gin.RouterGroup) {
	items := router.Group("/items")
	{
		// Public read‑only (auth required)
		items.GET("/", mw.Auth, FetchItems)
		items.GET("/:id", mw.Auth, FetchItemById)

		// Create (auth + user role)
		items.POST("/", mw.Auth, mw.CheckUserRole, CreateItem)

		// “My Items” — only the authenticated user’s items
		items.GET("/me", mw.Auth, mw.CheckUserRole, FetchAllItemsByUserId)

		// Update & delete — must be the owner
		// TODO: items.POST("/:id/image", mw.Auth, mw.CheckUserRole, UploadItemImage) // to update the uploaded image_original
		items.PATCH("/:id", mw.Auth, mw.CheckUserRole, UpdateItemById)
		items.PATCH("/:id/status", mw.Auth, mw.CheckUserRole, UpdateItemStatus)
		items.DELETE("/:id", mw.Auth, mw.CheckUserRole, DeleteItemById)
	}
}
