package context

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseParamUint(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error.invalid_id",
		})
		return 0, false
	}
	return uint(id), true
}
