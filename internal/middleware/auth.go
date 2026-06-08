package middleware

import "github.com/gin-gonic/gin"

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
		token := ctx.GetHeader("Authorization")

		if token == "" {
			ctx.AbortWithStatusJSON(401, gin.H{
				"error": "unauthorized",
			})
			return
		}

		ctx.Next()
	}
}
