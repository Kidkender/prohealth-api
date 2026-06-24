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

type HospitalHandler struct {
	service *service.HospitalService
}

func NewHospitalHandler(service *service.HospitalService) *HospitalHandler {
	return &HospitalHandler{service: service}
}

func (h *HospitalHandler) GetHospitals(c *gin.Context) {
	hospitals, err := h.service.GetHospitals(c)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToHospitalResponses(hospitals)))
}

func (h *HospitalHandler) CreateHospital(c *gin.Context) {
	var payload dto.CreateHospitalRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	if err := h.service.CreateHospital(c, payload); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *HospitalHandler) GetHospitalByID(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	hospital, err := h.service.GetHospitalById(c, id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(response.ToHospitalResponse(hospital)))
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
	if err := h.service.UpdateHospital(c, id, payload); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}

func (h *HospitalHandler) GetHospitalByMe(c *gin.Context) {
	userID := context.UserId(c)
	results, err := h.service.GetHospitalByUser(c, userID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	mapped := make([]response.MyHospitalResponse, 0, len(results))
	for _, r := range results {
		mapped = append(mapped, response.ToMyHospitalResponse(r.Hospital, r.Role))
	}
	c.JSON(http.StatusOK, response.Ok(mapped))
}
