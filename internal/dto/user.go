package dto

import "github.com/kidkender/prohealth-api/internal/model"

type AssignRoleRequest struct {
	Role   model.UserRole `json:"role" binding:"required,oneof=admin user doctor"`
	UserID uint           `json:"user_id"`
}
