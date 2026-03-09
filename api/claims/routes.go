package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func ClaimRoutes(router *gin.RouterGroup) {
	claims := router.Group("/claims")
	{
		// User operations
		claims.POST("/", mw.Auth, mw.CheckUserRole, CreateClaim)
		claims.POST("/:id/proof-image", mw.Auth, mw.CheckUserRole, UploadClaimProofImage)
		claims.GET("/me", mw.Auth, mw.CheckUserRole, FetchUserClaims)
		claims.GET("/item/:id", mw.Auth, FetchClaimsByItem)

		// Admin operations
		claims.GET("/admin", mw.Auth, mw.CheckAdminRole, FetchAllClaims)
		claims.PATCH("/:id", mw.Auth, mw.CheckAdminRole, ProcessClaim)
	}
}
