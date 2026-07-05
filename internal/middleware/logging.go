package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		attrs := []any{
			"request_id", c.GetString(RequestIDKey),
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}
		if hospitalID, ok := c.Get("hospital_id"); ok {
			attrs = append(attrs, "hospital_id", hospitalID)
		}

		if userID, ok := c.Get("user_id"); ok {
			attrs = append(attrs, "user_id", userID)
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "error", c.Errors.String())
			logger.Error("http_request", attrs...)
			return
		}

		logger.Info("http_request", attrs...)
	}
}
