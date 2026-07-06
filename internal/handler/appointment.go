package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/response"
	"github.com/kidkender/prohealth-api/internal/service"
)

type AppointmentHandler struct {
	appointmentService *service.AppointmentService
}

func NewAppointmentHandler(appointmentService *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointmentService: appointmentService}
}

func (h *AppointmentHandler) RegisterAppointment(ctx *gin.Context) {
	var req dto.CreateAppointmentInput
	hospitalID := context.HospitalId(ctx)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}
	if err := h.appointmentService.CreateAppointment(ctx.Request.Context(), req, hospitalID); err != nil {
		apperror.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, response.Done())
}

func (h *AppointmentHandler) GetAppointments(ctx *gin.Context) {
	var query dto.FilterAppointment
	if err := ctx.ShouldBindQuery(&query); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}
	appointments, total, err := h.appointmentService.GetAppointments(ctx.Request.Context(), context.HospitalId(ctx), query)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Paginated(appointments, total, query.Page, query.Limit))
}

func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAppointmentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	callerID := context.UserId(c)
	callerRole := model.UserRole(context.HospitalRole(c))
	if err := h.appointmentService.UpdateAppointment(c.Request.Context(), id, callerID, callerRole, req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *AppointmentHandler) GetAppointmentByMe(ctx *gin.Context) {
	userID := context.UserId(ctx)
	var filter dto.FilterAppointmentDoctor
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}
	appointments, err := h.appointmentService.GetAppointmentDoctor(ctx.Request.Context(), context.HospitalId(ctx), userID, &filter)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Ok(appointments))
}
