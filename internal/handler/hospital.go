package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/response"
	"github.com/kidkender/prohealth-api/internal/service"
)

type HospitalHandler struct {
	service *service.HospitalService
}

func NewHospitalHandler(service *service.HospitalService) *HospitalHandler {
	return &HospitalHandler{service: service}
}

func (h *HospitalHandler) GetHospitals(c *gin.Context) {
	hospitals, err := h.service.GetHospitals()
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	response.Ok(response.ToHospitalResponses(hospitals))
}

func (h *HospitalHandler) CreateHospital(c *gin.Context) {
	var payload dto.CreateHospitalRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.service.CreateHospital(payload); err != nil {
		apperror.HandleError(c, err)
		return
	}
	response.Done()
}

func (h *HospitalHandler) GetHospitalByID(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	hospital, err := h.service.GetHospitalById(id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	response.Ok(response.ToHospitalResponse(hospital))
}

func (h *HospitalHandler) UpdateHospital(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	var payload dto.UpdateHospitalRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.service.UpdateHospital(id, payload); err != nil {
		apperror.HandleError(c, err)
		return
	}
	response.Done()
}
