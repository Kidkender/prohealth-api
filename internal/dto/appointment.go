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
