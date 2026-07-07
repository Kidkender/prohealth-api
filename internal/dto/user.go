package dto

import "github.com/kidkender/prohealth-api/internal/model"

type AssignRoleRequest struct {
	Role   model.UserRole `json:"role" binding:"required,oneof=admin user doctor"`
	UserID uint           `json:"user_id"`
}

type FilterUser struct {
	PaginationQuery
	HospitalID   *uint             `form:"hospital_id"`
	Role         *model.UserRole   `form:"role"`
	DepartmentID *uint             `form:"department_id"`
	Status       *model.UserStatus `form:"status"`
	Search       *string           `form:"search"`
}

type UpdateUserStatusRequest struct {
	Status model.UserStatus `json:"status" binding:"required,oneof=active inactive banned"`
}
