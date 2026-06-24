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
}
