package pkg

import (
	"net/http"

	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/gin-gonic/gin"
)

type Validatable interface {
	Validate() (errorMsg string, err error)
}

func ValidateRequest[T Validatable](c *gin.Context) (*T, bool) {
	var req T

	if err := c.BindJSON(&req); err != nil {
		logger.Log.ErrorCtx(c, "[REQ-ERROR]: Failed to bind JSON", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return nil, false
	}

	if errorMsg, err := req.Validate(); err != nil {
		logger.Log.ErrorCtx(c, "[REQ-ERROR]: Validation failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": errorMsg})
		return nil, false
	}

	return &req, true
}
