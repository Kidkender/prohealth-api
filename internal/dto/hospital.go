package dto

type CreateHospitalRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Address     string `json:"address" binding:"required"`
}

type UpdateHospitalRequest struct {
	Description string `json:"description"`
	Address     string `json:"address"`
}
