package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userId := c.GetUint("user_id")

	result, err := h.userService.GetUser(c, userId)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"user":    result,
	})
}
