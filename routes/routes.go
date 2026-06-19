package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/container"
	"github.com/kidkender/prohealth-api/internal/handler"
)

func RegisterRoutes(r *gin.Engine, c *container.Container) {
	api := r.Group("/api/v1")

	authHandler := handler.NewAuthHandler(c.AuthService)
	userHandler := handler.NewUserHandler(c.UserService)
	scheduleHandler := handler.NewScheduleHandler(c.ScheduleService)
	appointmentHandler := handler.NewAppointmentHandler(c.AppointmentService)

	api.GET("/health", handler.HealthCheck)
	RegisterAuthRoutes(api, authHandler)
	RegisterUserRoutes(api, userHandler)
	RegisterScheduleRoutes(api, scheduleHandler)
	RegisterAppointmentRoutes(api, appointmentHandler)
}
