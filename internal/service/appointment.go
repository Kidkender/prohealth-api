package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/audit"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/event"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/outbox"
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
	db   *gorm.DB
	mail *MailService
	pub  *outbox.Publisher
}

func NewAppointmentService(db *gorm.DB, mail *MailService, pub *outbox.Publisher) *AppointmentService {
	return &AppointmentService{db: db, mail: mail, pub: pub}
}

func (s *AppointmentService) CreateAppointment(ctx context.Context, payload dto.CreateAppointmentInput, hospitalID uint) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, payload.ShiftID).Error; err != nil {
		return err
	}

	appointmentTime := payload.AppointmentTime
	if t, err := time.Parse(time.RFC3339, payload.AppointmentTime); err == nil {
		appointmentTime = t.Format("15:04")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		appointment := model.Appointment{
			ShiftID:         payload.ShiftID,
			HospitalID:      hospitalID,
			PatientName:     payload.PatientName,
			PatientPhone:    payload.PatientPhone,
			PatientEmail:    payload.PatientEmail,
			AppointmentTime: appointmentTime,
			Reason:          payload.Reason,
		}
		if err := tx.Create(&appointment).Error; err != nil {
			return err
		}
		if err := audit.Record(tx, hospitalID, nil, model.ActionCreate, "appointment", appointment.ID); err != nil {
			return err
		}
		return s.pub.Publish(tx, event.AppointmentCreated{
			AppointmentID: appointment.ID,
			ActorID:       nil,
			PatientName:   payload.PatientName,
			PatientEmail:  payload.PatientEmail,
			Reason:        string(payload.Reason),
			Date:          shift.Date.Format("2006-01-02"),
			Time:          appointmentTime,
		})
	})
}

func (s *AppointmentService) GetAppointments(ctx context.Context, hospitalID uint, filter dto.FilterAppointment) ([]AppointmentResult, int64, error) {
	var appointments []model.Appointment

	query := s.db.WithContext(ctx).Model(&model.Appointment{}).Where("hospital_id = ?", hospitalID)
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

func (s *AppointmentService) UpdateAppointment(ctx context.Context, id uint, callerID uint, callerRole model.UserRole, payload dto.UpdateAppointmentInput) error {
	var appt model.Appointment
	if err := s.db.WithContext(ctx).First(&appt, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrAppointmentNotFound
		}
		return err
	}

	if callerRole == model.RoleDoctor {
		var shift model.ScheduleShift
		err := s.db.WithContext(ctx).
			Joins("JOIN schedule_shift_doctors ON schedule_shift_doctors.shift_id = schedule_shifts.id").
			Where("schedule_shift_doctors.user_id = ? AND schedule_shifts.id = ?", callerID, appt.ShiftID).
			First(&shift).Error
		if err != nil {
			return apperror.ErrForbidden
		}

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

func (s *AppointmentService) GetAppointmentDoctor(ctx context.Context, hospitalID uint, doctorID uint, filter *dto.FilterAppointmentDoctor) ([]AppointmentResult, error) {
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
		Where("appointments.hospital_id = ?", hospitalID).
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
