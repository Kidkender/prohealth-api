package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RegisterAppointmentRoutes(router gin.IRouter, handler *handler.AppointmentHandler) {
	appointment := router.Group("/appointments")
	appointment.POST("/", handler.RegisterAppointment)
	appointment.GET("/", middleware.RequireRole(model.RoleAdmin), handler.GetAppointments)
	appointment.GET("/me", handler.GetAppointmentByMe)
	appointment.PATCH("/:id",
		middleware.RequireRole(model.RoleAdmin, model.RoleDoctor),
		handler.UpdateAppointment,
	)
}
