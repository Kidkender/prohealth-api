package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	err := h.authService.Register(c, &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": "ok",
	})

}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	result, err := h.authService.Login(c, &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result,
	})
}
