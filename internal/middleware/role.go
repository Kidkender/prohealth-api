package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := model.UserRole(context.UserRole(c))

		if slices.Contains(roles, role) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "forbidden",
		})
	}
}
