package dto

import "github.com/kidkender/prohealth-api/internal/model"

type InviteMemberRequest struct {
	Email string         `json:"email" binding:"required,email"`
	Role  model.UserRole `json:"role" binding:"required,oneof=admin doctor user"`
}

type AcceptInvitationRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Fullname string `json:"fullname" binding:"required"`
}
