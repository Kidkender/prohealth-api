package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
)

func RegisterUserRoutes(router gin.IRouter, userHandler *handler.UserHandler) {
	user := router.Group("/users")

	user.GET("/me", userHandler.GetMe)
}
