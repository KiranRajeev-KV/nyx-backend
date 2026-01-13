package pkg

import (
	"fmt"
	"net/http"

	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

func TagRequestWithId(c *gin.Context) {
	id := ksuid.New()
	c.Set("request_id", id.String())
	c.Next()
}

func GetEmail(c *gin.Context, route string) (string, bool) {
	email := c.GetString("email")
	if email == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		msg := fmt.Sprintf("[%s-FATAL]: No email after crossing auth middleware", route)
		logger.Log.FatalCtx(c, msg, nil)

		return "", false
	}

	return email, true
}
