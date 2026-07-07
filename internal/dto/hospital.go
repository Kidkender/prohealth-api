package dto

import "github.com/kidkender/prohealth-api/internal/model"

type AddHospitalMemberRequest struct {
	Email string         `json:"email" binding:"required,email"`
	Role  model.UserRole `json:"role" binding:"required,oneof=admin doctor user"`
}

type CreateHospitalRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Address     string `json:"address" binding:"required"`
}

type UpdateHospitalRequest struct {
	Description string `json:"description"`
	Address     string `json:"address"`
}

type UpdateHospitalStatusRequest struct {
	Status model.HospitalStatus `json:"status" binding:"required,oneof=approved rejected"`
}

type FilterHospitalStatus struct {
	Status *model.HospitalStatus `form:"status"`
}
