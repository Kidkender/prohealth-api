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

type ChatHandler struct {
	agentService *service.AgentService
}

func NewChatHandler(agentService *service.AgentService) *ChatHandler {
	return &ChatHandler{agentService: agentService}
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	hospitalID := context.HospitalId(c)
	resp, err := h.agentService.HandleMessage(c.Request.Context(), hospitalID, req.SessionID, nil, req.Message)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(resp))
}
