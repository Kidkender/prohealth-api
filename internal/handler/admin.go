package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/response"
	"github.com/kidkender/prohealth-api/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler(service *service.AdminService) *AdminHandler {
	return &AdminHandler{
		service: service,
	}
}

func (h *AdminHandler) GetAdminDashboard(ctx *gin.Context) {
	dashboard, err := h.service.GetSystemStat(ctx.Request.Context())
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Ok(dashboard))
}
