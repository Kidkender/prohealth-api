package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/timeutil"
	"gorm.io/gorm"
)

type AppointmentResult struct {
	ID              uint   `json:"id"`
	ShiftID         uint   `json:"shift_id"`
	ShiftDate       string `json:"shift_date"`
	PatientName     string `json:"patient_name"`
	PatientPhone    string `json:"patient_phone"`
	AppointmentTime string `json:"appointment_time"`
	Reason          string `json:"reason"`
	Status          string `json:"status"`
	Notes           string `json:"notes"`
}

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
	return s.db.Create(&appointment).Error
}

func (s *AppointmentService) GetAppointments(ctx context.Context, filter dto.FilterAppointment) ([]AppointmentResult, int64, error) {
	var appointments []model.Appointment

	query := s.db.WithContext(ctx).Model(&model.Appointment{})
	if filter.ShiftID != nil {
		query = query.Where("shift_id = ?", *filter.ShiftID)
	}
	if filter.PatientPhone != nil {
		query = query.Where("patient_phone = ?", *filter.PatientPhone)
	}
	if filter.Reason != nil {
		query = query.Where("reason = ?", *filter.Reason)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("created_at desc").
		Limit(filter.Limit).
		Offset(filter.Offset()).
		Preload("Shift").
		Find(&appointments).Error; err != nil {
		return nil, 0, err
	}

	results := make([]AppointmentResult, 0, len(appointments))
	for _, appointment := range appointments {
		results = append(results, AppointmentResult{
			ID:              appointment.ID,
			ShiftID:         appointment.ShiftID,
			ShiftDate:       appointment.Shift.Date.Format("2006-01-02"),
			PatientName:     appointment.PatientName,
			PatientPhone:    appointment.PatientPhone,
			AppointmentTime: appointment.AppointmentTime,
			Reason:          string(appointment.Reason),
			Status:          string(appointment.Status),
			Notes:           appointment.Notes,
		})
	}
	return results, count, nil
}

func (s *AppointmentService) UpdateAppointment(ctx context.Context, id uint, payload dto.UpdateAppointmentInput) error {
	var appt model.Appointment
	if err := s.db.WithContext(ctx).First(&appt, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrAppointmentNotFound
		}
		return err
	}
	if payload.Status != "" {
		appt.Status = payload.Status
	}
	if payload.Reason != "" {
		appt.Reason = payload.Reason
	}
	appt.Notes = payload.Notes
	return s.db.WithContext(ctx).Save(&appt).Error
}

func (s *AppointmentService) GetAppointmentDoctor(ctx context.Context, doctorID uint, filter *dto.FilterAppointmentDoctor) ([]AppointmentResult, error) {
	slog.Info("GetAppointmentDoctor", "doctorID", doctorID)

	if err := s.db.WithContext(ctx).First(&model.User{}, doctorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, err
	}

	from, to := timeutil.CurrentWeekRange()
	if filter.From != nil {
		from = *filter.From
	}
	if filter.To != nil {
		to = filter.To.Add(24 * time.Hour)
	}

	slog.Info("GetAppointmentDoctor", "from", from, "to", to)

	query := s.db.WithContext(ctx).
		Joins("JOIN schedule_shift_doctors ON schedule_shift_doctors.shift_id = appointments.shift_id").
		Joins("JOIN schedule_shifts ON schedule_shifts.id = appointments.shift_id").
		Where("schedule_shift_doctors.user_id = ?", doctorID).
		Where("schedule_shifts.date >= ? AND schedule_shifts.date < ?", from, to)

	var appointments []model.Appointment
	if err := query.Preload("Shift").Find(&appointments).Error; err != nil {
		return nil, err
	}

	result := make([]AppointmentResult, 0, len(appointments))
	for _, appointment := range appointments {
		result = append(result, AppointmentResult{
			ID:              appointment.ID,
			ShiftID:         appointment.ShiftID,
			ShiftDate:       appointment.Shift.Date.Format("2006-01-02"),
			PatientName:     appointment.PatientName,
			PatientPhone:    appointment.PatientPhone,
			AppointmentTime: appointment.AppointmentTime,
			Reason:          string(appointment.Reason),
			Status:          string(appointment.Status),
			Notes:           appointment.Notes,
		})
	}
	return result, nil
}
