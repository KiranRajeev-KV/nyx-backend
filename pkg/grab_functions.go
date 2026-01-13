package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
)
	
func GrabRequestId(c *gin.Context) string {
	reqId, ok := c.Get("request_id")
	if !ok {
		return "missing-id"
	}
	return fmt.Sprintf("%v", reqId)
}