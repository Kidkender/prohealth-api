package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

func TenantMiddleware(db *gorm.DB) gin.HandlerFunc {
	skipRoutes := map[skipKey]bool{
		{http.MethodPost, "/api/v1/auth/register"}: true,
		{http.MethodPost, "/api/v1/auth/login"}:    true,
		{http.MethodGet, "/api/v1/health"}:         true,
	}

	return func(c *gin.Context) {
		if skipRoutes[skipKey{c.Request.Method, c.FullPath()}] {
			c.Next()
			return
		}

		headerVal := c.GetHeader("X-Hospital-ID")
		if headerVal == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "missing X-Hospital-ID header",
			})
			return
		}

		hospitalID, err := strconv.ParseUint(headerVal, 10, 64)
		if err != nil || hospitalID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid X-Hospital-ID",
			})
			return
		}

		if userIDVal, exists := c.Get("user_id"); exists {
			if userID, ok := userIDVal.(uint); ok && userID > 0 {
				var member model.HospitalMember
				if err := db.
					Where("hospital_id = ? AND user_id = ?", hospitalID, userID).
					First(&member).Error; err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"error": "user does not belong to this hospital",
					})
					return
				}
			}
		}

		c.Set("hospital_id", uint(hospitalID))
		c.Next()
	}
}
