package dto

import (
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
)

type CreateAppointmentInput struct {
	ShiftID         uint                    `json:"shift_id" binding:"required"`
	PatientName     string                  `json:"patient_name" binding:"required"`
	PatientPhone    string                  `json:"patient_phone" binding:"required"`
	AppointmentTime time.Time               `json:"time" binding:"required"`
	Reason          model.AppointmentReason `json:"reason" binding:"required"`
}

type FilterAppointment struct {
	PaginationQuery
	ShiftID      *uint                    `form:"shift_id"`
	PatientPhone *string                  `form:"patient_phone"`
	Reason       *model.AppointmentReason `form:"reason"`
	Status       *model.AppointmentStatus `form:"status"`
}

type FilterAppointmentDoctor struct {
	From *time.Time `form:"from" time_format:"2006-01-02"`
	To   *time.Time `form:"to" time_format:"2006-01-02"`
}
