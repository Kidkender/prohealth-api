package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/service"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api/v1")

	authService := service.NewAuthService(db)
	authHandler := handler.NewAuthHandler(authService)

	api.GET("/health", handler.HealthCheck)
	RegisterAuthRoutes(api, authHandler)
}
