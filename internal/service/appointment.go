package service

import (
	"context"

	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type AppointmentService struct {
	db *gorm.DB
}

func NewAppointmentService(db *gorm.DB) *AppointmentService {
	return &AppointmentService{db: db}
}

func (s *AppointmentService) CreateAppointment(ctx context.Context, payload dto.CreateAppointmentInput) error {
	var shift model.ScheduleShift
	if err := s.db.First(&shift, payload.ShiftID).Error; err != nil {
		return err
	}

	appointment := model.Appointment{
		ShiftID:         payload.ShiftID,
		PatientName:     payload.PatientName,
		PatientPhone:    payload.PatientPhone,
		AppointmentTime: payload.AppointmentTime,
		Reason:          payload.Reason,
	}
	if err := s.db.Create(&appointment).Error; err != nil {
		return err
	}

	return nil
}
