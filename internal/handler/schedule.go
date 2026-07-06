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

type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

func (h *ScheduleHandler) CreatePeriod(c *gin.Context) {
	var req dto.CreateScheduleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.scheduleService.CreateSchedulePeriod(c.Request.Context(), context.HospitalId(c), req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Done())
}

func (h *ScheduleHandler) CreateShift(c *gin.Context) {
	var req dto.CreateScheduleShiftInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.scheduleService.CreateScheduleShift(c.Request.Context(), context.HospitalId(c), req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Done())
}

func (h *ScheduleHandler) SetDoctorShift(c *gin.Context) {
	var req dto.SetDoctorForShiftInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.scheduleService.SetDoctorForShift(c.Request.Context(), req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
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
	if err := h.scheduleService.UpdateSchedulePeriod(c.Request.Context(), uint(id), req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
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
	var hospitalID uint
	if filter.HospitalID != nil {
		hospitalID = *filter.HospitalID
	} else if id, exists := c.Get("hospital_id"); exists {
		hospitalID, _ = id.(uint)
	}
	periods, total, err := h.scheduleService.GetSchedulePeriods(c.Request.Context(), hospitalID, filter)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Paginated(
		response.ToSchedulePeriodResponses(periods), total, filter.Page, filter.Limit,
	))
}

func (h *ScheduleHandler) GetSchedulePeriodById(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	period, err := h.scheduleService.GetSchedulePeriodById(c.Request.Context(), uint(id))
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToSchedulePeriodResponse(*period)))
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
	shifts, total, err := h.scheduleService.GetScheduleShifts(c.Request.Context(), context.HospitalId(c), filter)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Paginated(
		response.ToScheduleShiftResponses(shifts), total, filter.Page, filter.Limit,
	))
}

func (h *ScheduleHandler) GetShiftsByPeriod(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	shifts, err := h.scheduleService.GetScheduleShiftsByPeriodID(c.Request.Context(), uint(id))
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToScheduleShiftResponses(shifts)))
}

func (h *ScheduleHandler) GetScheduleShiftByID(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	shift, err := h.scheduleService.GetScheduleShiftById(c.Request.Context(), uint(id))
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToScheduleShiftResponse(*shift)))
}

func (h *ScheduleHandler) DeleteSchedulePeriod(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	if err := h.scheduleService.DeleteSchedulePeriod(c.Request.Context(), id); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *ScheduleHandler) DeleteScheduleShift(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	if err := h.scheduleService.DeleteScheduleShift(c.Request.Context(), id); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *ScheduleHandler) GenerateSchedule(c *gin.Context) {
	var req dto.GenerateScheduleInput
	slog.Info("GenerateSchedule", "req", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		slog.Error("GenerateSchedule", "err", err)
		return
	}
	slog.Info("GenerateSchedule 2", "req", req)
	result, err := h.scheduleService.GenerateSchedule(c.Request.Context(), context.HospitalId(c), req)
	if err != nil {
		apperror.HandleError(c, err)
		slog.Error("GenerateSchedule 2", "err", err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(result))
}

func (h *ScheduleHandler) ReplaceDoctorsForShift(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	var payload struct {
		DoctorIDs []uint `json:"doctor_ids"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.scheduleService.ReplaceDoctorsForShift(c.Request.Context(), uint(id), payload.DoctorIDs); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *ScheduleHandler) UpdateScheduleShift(c *gin.Context) {
	var req dto.UpdateScheduleShiftInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	if err := h.scheduleService.UpdateScheduleShift(c.Request.Context(), uint(id), req); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *ScheduleHandler) GetScheduleShiftByDoctor(c *gin.Context) {
	var filter dto.FilterScheduleShift
	if err := c.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	result, err := h.scheduleService.GetScheduleShiftByDoctor(c.Request.Context(), context.HospitalId(c), uint(id), &filter)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToScheduleShiftResponses(result)))
}
