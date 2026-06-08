package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	api.GET("/health", handler.HealthCheck)
	RegisterAuthRoutes(api, &handler.AuthHandler{})
}
