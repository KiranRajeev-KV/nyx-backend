package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

func TagRequestWithId(c *gin.Context) {
	id := ksuid.New()
	c.Set("request_id", id.String())
	c.Next()
}
