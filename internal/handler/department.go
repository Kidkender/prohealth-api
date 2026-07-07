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

type DepartmentHandler struct {
	service *service.DepartmentService
}

func NewDepartmentHandler(service *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var payload dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	department, err := h.service.CreateDepartment(c.Request.Context(), payload)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(department))
}

func (h *DepartmentHandler) GetDepartments(c *gin.Context) {
	departments, err := h.service.GetDepartments(c.Request.Context())
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(departments))
}

func (h *DepartmentHandler) GetDepartmentByID(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	department, err := h.service.GetDepartmentByID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(department))
}

func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	var payload dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		apperror.ValidationError(c, err)
		return
	}
	department, err := h.service.UpdateDepartment(c.Request.Context(), id, payload)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Ok(department))
}

func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id, ok := context.ParseParamUint(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteDepartment(c.Request.Context(), id); err != nil {
		apperror.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Done())
}
