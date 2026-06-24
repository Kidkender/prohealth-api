package response

import "github.com/kidkender/prohealth-api/internal/model"

type HospitalResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

type MyHospitalResponse struct {
	Hospital HospitalResponse `json:"hospital"`
	Role     model.UserRole   `json:"role"`
}

func ToHospitalResponse(hospital *model.Hospital) HospitalResponse {
	return HospitalResponse{
		ID:          hospital.ID,
		Name:        hospital.Name,
		Description: hospital.Description,
		Slug:        hospital.Slug,
	}
}

func ToHospitalResponses(hospitals []model.Hospital) []HospitalResponse {
	responses := make([]HospitalResponse, 0, len(hospitals))
	for _, hospital := range hospitals {
		responses = append(responses, ToHospitalResponse(&hospital))
	}
	return responses
}

func ToMyHospitalResponse(h model.Hospital, role model.UserRole) MyHospitalResponse {
	return MyHospitalResponse{
		Hospital: ToHospitalResponse(&h),
		Role:     role,
	}
}
