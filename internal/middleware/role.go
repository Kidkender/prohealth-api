package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := model.UserRole(context.HospitalRole(c))
		if role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrForbidden,
			})
			return
		}

		if slices.Contains(roles, role) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "forbidden",
		})
	}
}

func RequireSystemRole(roles ...model.SystemRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := model.SystemRole(context.SystemRole(c))
		if role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrForbidden,
			})
			return
		}

		if slices.Contains(roles, role) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "forbidden",
		})
	}
}
