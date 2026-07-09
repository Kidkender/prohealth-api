package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/security"
)

type skipKey struct {
	method string
	path   string
}

func AuthMiddleware() gin.HandlerFunc {
	skipRoutes := map[skipKey]bool{
		{http.MethodPost, "/api/v1/auth/register"}:            true,
		{http.MethodPost, "/api/v1/auth/login"}:               true,
		{http.MethodGet, "/api/v1/health"}:                    true,
		{http.MethodGet, "/api/v1/schedule/period"}:           true,
		{http.MethodGet, "/api/v1/schedule/period/:id/shift"}: true,
		{http.MethodPost, "/api/v1/appointments/"}:            true,
		{http.MethodGet, "/api/v1/hospital"}:                  true,
		{http.MethodGet, "/api/v1/hospital/:id"}:              true,
		{http.MethodPost, "/api/v1/chat"}:                     true,
	}

	return func(ctx *gin.Context) {

		key := skipKey{method: ctx.Request.Method, path: ctx.FullPath()}
		if skipRoutes[key] {
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
				"error": apperror.ErrInvalidToken,
			})
			return
		}
		slog.Info("Authenticate claims", "claims", claims)

		ctx.Set("user_id", claims.UserId)
		ctx.Set("system_role", string(claims.SystemRole))
		ctx.Next()
	}
}
