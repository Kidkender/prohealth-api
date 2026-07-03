package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/ratelimit"
)

func RegisterAuthRoutes(router gin.IRouter, authHandler *handler.AuthHandler, authLimiter ratelimit.Limiter) {
	auth := router.Group("/auth")
	auth.Use(middleware.RateLimitMiddleware(authLimiter, middleware.BuildIPKey))
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.PUT("/change-password", authHandler.ChangePassword)
	auth.POST("/refresh-token", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)
}
