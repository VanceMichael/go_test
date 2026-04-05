package middleware

import (
	"time"

	"release-manager/internal/config"

	"github.com/gin-gonic/gin"
)

func Logger(logger *config.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		logger.Infow("HTTP Request",
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"ip", c.ClientIP(),
			"latency", latency.String(),
			"user-agent", c.Request.UserAgent(),
		)
	}
}
