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

type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

func (h *ScheduleHandler) CreatePeriod(c *gin.Context) {
	var req dto.CreateScheduleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	err := h.scheduleService.CreateSchedulePeriod(c, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success": "ok",
	})
}

func (h *ScheduleHandler) CreateShift(c *gin.Context) {
	var req dto.CreateScheduleShiftInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	err := h.scheduleService.CreateScheduleShift(c, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success": "ok",
	})
}

func (h *ScheduleHandler) SetDoctorShift(c *gin.Context) {
	var req dto.SetDoctorForShiftInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	err := h.scheduleService.SetDoctorForShift(c, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": "ok",
	})
}

func (h *ScheduleHandler) UpdatePeriod(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSchedulePeriodInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	if err := h.scheduleService.UpdateSchedulePeriod(c, uint(id), req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": "ok",
	})
}

func (h *ScheduleHandler) GetSchedulePeriods(c *gin.Context) {
	var filter dto.FilterSchedulePeriod
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	filter.Normalize()
	if err := filter.Validate(); err != nil {
		apperror.ValidationError(c, err)
		return
	}

	periods, total, err := h.scheduleService.GetSchedulePeriods(c, filter)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(
		http.StatusOK,
		response.NewPaginatedResponse(
			response.ToSchedulePeriodResponses(periods), total, filter.Page, filter.Limit),
	)
}

func (h *ScheduleHandler) GetSchedulePeriodById(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	period, err := h.scheduleService.GetSchedulePeriodById(c, uint(id))
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.ToSchedulePeriodResponse(*period))
}

func (h *ScheduleHandler) GetAllShifts(c *gin.Context) {
	var filter dto.FilterScheduleShift
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	filter.Normalize()

	if err := filter.Validate(); err != nil {
		apperror.HandleError(c, err)
		return
	}
	shifts, total, err := h.scheduleService.GetScheduleShifts(c, filter)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(
		http.StatusOK,
		response.NewPaginatedResponse(shifts, total, filter.Page, filter.Limit),
	)
}

func (h *ScheduleHandler) GetShiftsByPeriod(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	shifts, err := h.scheduleService.GetScheduleShiftsByPeriodID(c, uint(id))
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(
		http.StatusOK,
		response.ToScheduleShiftResponses(shifts),
	)
}

func (h *ScheduleHandler) GetScheduleShiftByID(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	shift, err := h.scheduleService.GetScheduleShiftById(c, uint(id))
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.ToScheduleShiftResponse(*shift))
}
