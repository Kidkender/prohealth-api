package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/context"
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

	result, err := h.userService.GetUser(c, userId)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(result))
}
