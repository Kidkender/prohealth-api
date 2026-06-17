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
	userService := service.NewUserService(db)
	userHandler := handler.NewUserHandler(userService)
	scheduleService := service.NewScheduleService(db)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	appointmentService := service.NewAppointmentService(db)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService)

	api.GET("/health", handler.HealthCheck)
	RegisterAuthRoutes(api, authHandler)
	RegisterUserRoutes(api, userHandler)
	RegisterScheduleRoutes(api, scheduleHandler)
	RegisterAppointmentRoutes(api, appointmentHandler)
}
