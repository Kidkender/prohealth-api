package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/service"
)

func TenantMiddleware(hospitalService *service.HospitalService) gin.HandlerFunc {
	skipRoutes := map[skipKey]bool{
		{http.MethodPost, "/api/v1/auth/register"}:            true,
		{http.MethodPost, "/api/v1/auth/login"}:               true,
		{http.MethodPut, "/api/v1/auth/change-password"}:      true,
		{http.MethodGet, "/api/v1/health"}:                    true,
		{http.MethodGet, "/api/v1/users/me"}:                  true,
		{http.MethodGet, "/api/v1/hospital/me"}:               true,
		{http.MethodGet, "/api/v1/hospital"}:                  true,
		{http.MethodPost, "/api/v1/hospital"}:                 true,
		{http.MethodGet, "/api/v1/hospital/:id"}:              true,
		{http.MethodGet, "/api/v1/schedule/period"}:           true,
		{http.MethodGet, "/api/v1/schedule/period/:id/shift"}: true,
		{http.MethodPost, "/api/v1/appointments/"}:            true,
		{http.MethodGet, "/api/v1/admin/stats"}:               true,
		{http.MethodGet, "/api/v1/admin/users"}:               true,
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

		var role string
		if userIDVal, exists := c.Get("user_id"); exists {
			if userID, ok := userIDVal.(uint); ok && userID > 0 {
				memberRole, err := hospitalService.GetMemberRole(c, uint(hospitalID), userID)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"error": apperror.ErrForbidden,
					})
					return
				}
				role = string(memberRole)
			}
		}

		c.Set("hospital_id", uint(hospitalID))
		c.Set("hospital_role", role)
		c.Next()
	}
}
