package middleware

import (
	"github.com/gin-gonic/gin"
)

func BuildIPKey(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}
	return ip
}
