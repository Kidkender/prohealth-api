package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
)

func RegisterAuthRoutes(router gin.IRouter, authHandler *handler.AuthHandler) {
	auth := router.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.PUT("/change-password", authHandler.ChangePassword)
}
