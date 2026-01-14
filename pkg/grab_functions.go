package pkg

import (
	"fmt"
	"net/http"

	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func GrabUserId(c *gin.Context, route string) (string, bool) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Oops! Something happened. Please try again later.",
		})
		msg := fmt.Sprintf("[%s-FATAL]: Missing userId in context after auth middleware", route)
		logger.Log.FatalCtx(c, msg, nil)

		return "", false
	}
	return userId, true
}

func GrabUuid(c *gin.Context, uuidStr string, route string, entity string) (uuid.UUID, bool) {
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "The request is malformed.",
		})
		msg := fmt.Sprintf("[%s-ERROR]: Failed to parse %s UUID", route, entity)
		logger.Log.ErrorCtx(c, msg, err)

		return uuid, false
	}
	return uuid, true
}
