package context

import "github.com/gin-gonic/gin"

func UserId(c *gin.Context) uint {
	return c.MustGet("user_id").(uint)
}

func HospitalId(c *gin.Context) uint {
	return c.MustGet("hospital_id").(uint)
}

func HospitalRole(c *gin.Context) string {
	return c.GetString("hospital_role")
}

func SystemRole(c *gin.Context) string {
	return c.GetString("system_role")
}
