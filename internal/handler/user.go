package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/response"
	"github.com/kidkender/prohealth-api/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userId := context.UserId(c)
	hospitalID := context.HospitalId(c)

	result, err := h.userService.GetUser(c, userId, hospitalID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(result))
}

func (h *UserHandler) AssignRole(c *gin.Context) {
	var payload dto.AssignRoleRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	hospitalID := context.HospitalId(c)
	if err := h.userService.AssignRole(c, hospitalID, payload); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}
