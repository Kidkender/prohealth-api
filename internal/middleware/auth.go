package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/security"
)

func AuthMiddleware() gin.HandlerFunc {

	skipPaths := map[string]bool{
		"/api/v1/auth/register": true,
		"/api/v1/auth/login":    true,
		"/api/v1/health":        true,
	}

	return func(ctx *gin.Context) {

		if skipPaths[ctx.FullPath()] {
			ctx.Next()
			return
		}
		header := ctx.GetHeader("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "error.auth.missing_token",
			})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := security.ParseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "error.auth.invalid_token",
			})
			return
		}

		ctx.Set("user_id", claims.UserId)
		ctx.Set("role", string(claims.Role))
		ctx.Next()
	}
}
