package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RegisterDepartmentRoutes(router gin.IRouter, handler *handler.DepartmentHandler) {
	department := router.Group("/departments")
	department.GET("", handler.GetDepartments)
	department.GET("/:id", handler.GetDepartmentByID)
	department.POST("", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.CreateDepartment)
	department.PUT("/:id", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.UpdateDepartment)
	department.DELETE("/:id", middleware.RequireSystemRole(model.SystemRoleAdmin), handler.DeleteDepartment)
}
