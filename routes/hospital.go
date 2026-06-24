package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RegisterHospitalRoutes(router gin.IRouter, handler *handler.HospitalHandler) {
	hospital := router.Group("/hospital")
	hospital.GET("", handler.GetHospitals)
	hospital.POST("", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.CreateHospital)
	hospital.GET("/me", handler.GetHospitalByMe)
	hospital.GET("/:id", handler.GetHospitalByID)
	hospital.PUT("/:id", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.UpdateHospital)
}
