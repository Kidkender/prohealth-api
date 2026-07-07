package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RegisterAdminRoutes(r gin.IRouter, h *handler.AdminHandler) {
	admin := r.Group("/admin", middleware.RequireSystemRole(model.SystemRoleAdmin))
	admin.GET("/stats", h.GetAdminDashboard)
	admin.GET("/users", h.GetUsers)
	admin.GET("/users/:id", h.GetUserDetail)
	admin.PATCH("/users/:id/status", h.UpdateUserStatus)
	admin.GET("/users/:id/activity", h.GetUserActivity)
	admin.GET("/hospitals", h.ListHospitals)
	admin.PATCH("/hospitals/:id/status", h.UpdateHospitalStatus)
	admin.GET("/hospitals/:id/members", h.GetHospitalMembers)
	admin.POST("/hospitals/:id/members", h.AddHospitalMember)
	admin.DELETE("/hospitals/:id/members/:userId", h.RemoveHospitalMember)
}
