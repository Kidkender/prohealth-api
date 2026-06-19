package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
)

func RegisterAppointmentRoutes(router gin.IRouter, handler *handler.AppointmentHandler) {
	appointment := router.Group("/appointments")
	appointment.POST("/", handler.RegisterAppointment)
	appointment.GET("/", handler.GetAppointments)
	appointment.GET("/me", handler.GetAppointmentByMe)
	appointment.PATCH("/:id", handler.UpdateAppointment)
}
