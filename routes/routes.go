package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/container"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/ratelimit"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(
	r *gin.Engine,
	c *container.Container,
	authLimiter, apiLimiter *ratelimit.TokenBucket,
	redisClient *redis.Client,
) {
	api := r.Group("/api/v1")

	api.Use(middleware.TimeoutMiddleware(5 * time.Second))
	api.GET("/health", handler.HealthCheck)

	api.Use(middleware.RateLimitMiddleware(apiLimiter, middleware.BuildIPKey))

	authHandler := handler.NewAuthHandler(c.AuthService)
	userHandler := handler.NewUserHandler(c.UserService)
	scheduleHandler := handler.NewScheduleHandler(c.ScheduleService)
	appointmentHandler := handler.NewAppointmentHandler(c.AppointmentService)
	hospitalHandler := handler.NewHospitalHandler(c.HospitalService)
	adminHandler := handler.NewAdminHandler(c.AdminService)

	RegisterAuthRoutes(api, authHandler, authLimiter)
	RegisterUserRoutes(api, userHandler)
	RegisterScheduleRoutes(api, scheduleHandler)
	RegisterAppointmentRoutes(api, appointmentHandler, redisClient)
	RegisterHospitalRoutes(api, hospitalHandler)
	RegisterAdminRoutes(api, adminHandler)
}
