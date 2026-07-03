package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/response"
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

	c.JSON(http.StatusCreated, response.Done())

}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	slog.Info("login request", "email", req.Email)

	result, err := h.authService.Login(c, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Ok(result))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	userID := context.UserId(c)
	err := h.authService.ChangePassword(c, userID, &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Done())
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	result, err := h.authService.RefreshToken(c, req.RefreshToken)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Ok(result))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	if err := h.authService.Logout(c, req.RefreshToken); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Done())
}
