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

type InvitationHandler struct {
	service *service.InvitationService
}

func NewInvitationHandler(service *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{service: service}
}

func (h *InvitationHandler) Invite(c *gin.Context) {
	var payload dto.InviteMemberRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	hospitalID := context.HospitalId(c)
	actorID := context.UserId(c)
	if err := h.service.Invite(c.Request.Context(), actorID, hospitalID, payload.Email, payload.Role); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *InvitationHandler) ListPending(c *gin.Context) {
	hospitalID := context.HospitalId(c)
	invitations, err := h.service.ListPending(c.Request.Context(), hospitalID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToPendingInvitationResponses(invitations)))
}

func (h *InvitationHandler) Revoke(c *gin.Context) {
	invitationID, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}

	hospitalID := context.HospitalId(c)
	actorID := context.UserId(c)
	if err := h.service.Revoke(c.Request.Context(), actorID, hospitalID, invitationID); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	var payload dto.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	if err := h.service.AcceptInvitation(c.Request.Context(), payload.Token, payload.Password, payload.Fullname); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}
