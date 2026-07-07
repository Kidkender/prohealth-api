package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/apperror"
	appcontext "github.com/kidkender/prohealth-api/internal/context"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/response"
	"github.com/kidkender/prohealth-api/internal/service"
)

type AdminHandler struct {
	service         *service.AdminService
	userService     *service.UserService
	auditService    *service.AuditService
	hospitalService *service.HospitalService
}

func NewAdminHandler(
	service *service.AdminService,
	userService *service.UserService,
	auditService *service.AuditService,
	hospitalService *service.HospitalService,
) *AdminHandler {
	return &AdminHandler{
		service:         service,
		userService:     userService,
		auditService:    auditService,
		hospitalService: hospitalService,
	}
}

func (h *AdminHandler) GetAdminDashboard(ctx *gin.Context) {
	dashboard, err := h.service.GetSystemStat(ctx.Request.Context())
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Ok(dashboard))
}

func (h *AdminHandler) GetUsers(ctx *gin.Context) {
	var filter dto.FilterUser
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}
	filter.Normalize()
	if err := filter.Validate(); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}
	users, count, err := h.userService.GetUsers(ctx.Request.Context(), filter)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, response.Paginated(users, count, filter.Page, filter.Limit))
}

func (h *AdminHandler) GetUserDetail(ctx *gin.Context) {
	targetUserID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}

	detail, err := h.userService.GetUserDetail(ctx.Request.Context(), targetUserID)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Ok(detail))
}

func (h *AdminHandler) UpdateUserStatus(ctx *gin.Context) {
	targetUserID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}

	actorID := appcontext.UserId(ctx)
	if err := h.userService.UpdateStatus(ctx.Request.Context(), actorID, targetUserID, req.Status); err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Ok(gin.H{"id": targetUserID, "status": req.Status}))
}

func (h *AdminHandler) ListHospitals(ctx *gin.Context) {
	var filter dto.FilterHospitalStatus
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}

	hospitals, err := h.hospitalService.GetHospitalsByStatus(ctx.Request.Context(), filter.Status)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Ok(hospitals))
}

func (h *AdminHandler) UpdateHospitalStatus(ctx *gin.Context) {
	hospitalID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}

	var req dto.UpdateHospitalStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}

	if err := h.hospitalService.UpdateHospitalStatus(ctx.Request.Context(), hospitalID, req.Status); err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Ok(gin.H{"id": hospitalID, "status": req.Status}))
}

func (h *AdminHandler) GetHospitalMembers(ctx *gin.Context) {
	hospitalID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}

	members, err := h.hospitalService.GetHospitalMembers(ctx.Request.Context(), hospitalID)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Ok(members))
}

func (h *AdminHandler) AddHospitalMember(ctx *gin.Context) {
	hospitalID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}

	var req dto.AddHospitalMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}

	if err := h.hospitalService.AddMemberByEmail(ctx.Request.Context(), hospitalID, req.Email, req.Role); err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Done())
}

func (h *AdminHandler) RemoveHospitalMember(ctx *gin.Context) {
	hospitalID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}
	userID, ok := appcontext.ParseParamUint(ctx, "userId")
	if !ok {
		return
	}

	if err := h.hospitalService.RemoveMember(ctx.Request.Context(), hospitalID, userID); err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Done())
}

func (h *AdminHandler) GetUserActivity(ctx *gin.Context) {
	targetUserID, ok := appcontext.ParseParamUint(ctx, "id")
	if !ok {
		return
	}

	var filter dto.FilterAuditLog
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}
	filter.Normalize()
	if err := filter.Validate(); err != nil {
		apperror.ValidationError(ctx, err)
		return
	}

	logs, count, err := h.auditService.GetUserActivity(ctx.Request.Context(), targetUserID, filter)
	if err != nil {
		apperror.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Paginated(logs, count, filter.Page, filter.Limit))
}
