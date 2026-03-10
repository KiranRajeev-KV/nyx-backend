package api

import (
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func AuditRoutes(router *gin.RouterGroup) {
	audit := router.Group("/audit")
	{
		audit.GET("/logs", mw.Auth, mw.CheckAdminRole, FetchAuditLogs)
	}
}
