package response

import (
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
)

type SchedulePeriodResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func ToSchedulePeriodResponse(p model.SchedulePeriod) SchedulePeriodResponse {
	return SchedulePeriodResponse{
		ID:        p.ID,
		Name:      p.Name,
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
	}
}

func ToSchedulePeriodResponses(ps []model.SchedulePeriod) []SchedulePeriodResponse {
	responses := make([]SchedulePeriodResponse, 0, len(ps))
	for _, p := range ps {
		responses = append(responses, ToSchedulePeriodResponse(p))
	}
	return responses
}

type ScheduleShiftResponse struct {
	ID             uint      `json:"id"`
	PeriodID       uint      `json:"period_id"`
	DepartmentID   uint      `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	Date           time.Time `json:"date"`
	TimeStart      string    `json:"time_start"`
	TimeEnd        string    `json:"time_end"`
	RoomNumber     string    `json:"room_number"`
	Status         string    `json:"status"`
	Note           string    `json:"note"`
	Doctors        []string  `json:"doctors"`
	CreatedAt      time.Time `json:"created_at"`
}

func ToScheduleShiftResponse(s model.ScheduleShift) ScheduleShiftResponse {
	doctors := make([]string, 0, len(s.Doctors))
	for _, d := range s.Doctors {
		doctors = append(doctors, d.UserDetail.Fullname)
	}

	deptName := ""
	if s.Department.ID != 0 {
		deptName = s.Department.Name
	}

	return ScheduleShiftResponse{
		ID:             s.ID,
		PeriodID:       s.PeriodID,
		DepartmentID:   s.DepartmentID,
		DepartmentName: deptName,
		Date:           s.Date,
		TimeStart:      s.TimeStart,
		TimeEnd:        s.TimeEnd,
		RoomNumber:     s.RoomNumber,
		Status:         string(s.Status),
		Note:           s.Note,
		Doctors:        doctors,
		CreatedAt:      s.CreatedAt,
	}
}

func ToScheduleShiftResponses(shifts []model.ScheduleShift) []ScheduleShiftResponse {
	responses := make([]ScheduleShiftResponse, 0, len(shifts))
	for _, s := range shifts {
		responses = append(responses, ToScheduleShiftResponse(s))
	}
	return responses
}
