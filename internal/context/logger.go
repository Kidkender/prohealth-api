package context

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Logger(c *gin.Context, base *slog.Logger) *slog.Logger {
	l := base.With("request_id", c.GetString("request_id"))
	if hospitalID, ok := c.Get("hospital_id"); ok {
		l = l.With("hospital_id", hospitalID)
	}
	if userID, ok := c.Get("user_id"); ok {
		l = l.With("user_id", userID)
	}
	return l
}
