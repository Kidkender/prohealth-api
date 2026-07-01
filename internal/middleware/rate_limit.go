package middleware

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/ratelimit"
)

func RateLimitMiddleware(
	limiter ratelimit.Limiter,
	keyFN func(*gin.Context) string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFN(c)
		if key == "" {
			c.Next()
			return
		}
		result, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			slog.Error("limiter error: ", "err", err)
			c.Next()
			return
		}
		c.Header("X-RateLimit-Remaining", strconv.FormatFloat(float64(result.Remaining), 'f', 0, 64))
		if !result.Allowed {
			c.Error(apperror.ErrRateLimitExceeded)
			return
		}
		c.Next()
	}
}
