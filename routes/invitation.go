package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RegisterInvitationRoutes(router gin.IRouter, handler *handler.InvitationHandler) {
	invitations := router.Group("/hospital/admin/invitations")
	invitations.POST("", middleware.RequireRole(model.RoleAdmin), handler.Invite)
	invitations.GET("", middleware.RequireRole(model.RoleAdmin), handler.ListPending)
	invitations.DELETE("/:id", middleware.RequireRole(model.RoleAdmin), handler.Revoke)

	router.POST("/invitations/accept", handler.Accept)
}
