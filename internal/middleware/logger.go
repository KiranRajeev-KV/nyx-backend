package mw

import (
	"fmt"
	"strings"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func LogMiddleware(l *logger.LoggerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		pathParams := map[string]string{}
		for _, param := range c.Params {
			pathParams[param.Key] = param.Value
		}
		pathParamsStr := fmt.Sprintf("%v", pathParams)

		queryParams := make(map[string]string)
		for k, v := range c.Request.URL.Query() {
			if len(v) == 1 {
				queryParams[k] = v[0]
			} else {
				queryParams[k] = strings.Join(v, ",")
			}
		}
		queryParamsStr := fmt.Sprintf("%v", queryParams)

		l.Logger.WithLevel(zerolog.InfoLevel).
			Str("req-id", logger.GrabRequestId(c)).
			Str("route", c.FullPath()).
			Str("method", c.Request.Method).
			Int("status", c.Writer.Status()).
			Int("response-size", c.Writer.Size()).
			Dur("duration", time.Since(start)).
			Str("path-params", pathParamsStr).
			Str("query-params", queryParamsStr).
			Str("ip", c.ClientIP()).
			Str("user-agent", c.Request.UserAgent())
	}
}
