package response

type UserResult struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	HospitalID     uint   `json:"hospital_id"`
	DepartmentName string `json:"department_name"`
	Address        string `json:"address"`
	PhoneNumber    string `json:"phone_number"`
}
