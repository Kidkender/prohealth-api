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
	hospital.POST("", handler.CreateHospital)
	hospital.GET("/me", handler.GetHospitalByMe)
	hospital.GET("/members", middleware.RequireRole(model.RoleAdmin), handler.GetMembers)
	hospital.POST("/members", middleware.RequireRole(model.RoleAdmin), handler.AddMember)
	hospital.DELETE("/members/:userId", middleware.RequireRole(model.RoleAdmin), handler.RemoveMember)
	hospital.GET("/:id/doctors", handler.GetHospitalDoctors)
	hospital.GET("/:id", handler.GetHospitalByID)
	hospital.PUT("/:id", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.UpdateHospital)
	hospital.DELETE("/:id", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.DeleteHospital)

	hospital.GET("/admin/stats", handler.GetHospitalStats, middleware.RequireRole(model.RoleAdmin))
	hospital.GET("/admin/appointment-trends", middleware.RequireRole(model.RoleAdmin), handler.GetAppointmentTrends)
	hospital.GET("/admin/department-performance", middleware.RequireRole(model.RoleAdmin), handler.GetDepartmentPerformance)
}
