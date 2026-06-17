package context

import "github.com/gin-gonic/gin"

func UserId(c *gin.Context) uint {
	return c.MustGet("user_id").(uint)
}

func UserRole(c *gin.Context) string {
	return c.GetString("role")
}
