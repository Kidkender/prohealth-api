package dto

type RegisterRequest struct {
	HospitalID uint   `json:"hospital_id"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	Fullname   string `json:"fullname"`
	Address    string `json:"address"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"oldPassword" binding:"required,min=8"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
