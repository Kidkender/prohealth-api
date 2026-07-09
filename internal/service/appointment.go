package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/audit"
	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/event"
	"github.com/kidkender/prohealth-api/internal/lock"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/outbox"
	"github.com/kidkender/prohealth-api/internal/timeutil"
	"gorm.io/gorm"
)

const appointmentDuration = 30 * time.Minute

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
	lock *lock.DistributedLock
}

func NewAppointmentService(
	db *gorm.DB,
	mail *MailService,
	pub *outbox.Publisher,
	lock *lock.DistributedLock,
) *AppointmentService {
	return &AppointmentService{db: db, mail: mail, pub: pub, lock: lock}
}

func (s *AppointmentService) CreateAppointment(ctx context.Context, payload dto.CreateAppointmentInput, hospitalID uint) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).
		Joins("JOIN schedule_shift_doctors ON schedule_shift_doctors.shift_id = schedule_shifts.id").
		Preload("Doctors.UserDetail").
		First(&shift, payload.ShiftID).Error; err != nil {
		return err
	}

	if hospitalID != shift.HospitalID {
		slog.Error("create_appointment", "hospital_id", hospitalID, "shift_hospital_id", shift.HospitalID)
		return apperror.ErrShiftNotFound
	}

	if shift.Status != model.ShiftScheduled {
		slog.Error("create_appointment", "shift_status", shift.Status)
		return apperror.ErrShiftNotFound
	}

	if len(shift.Doctors) == 0 {
		slog.Error("create_appointment", "doctors_count", len(shift.Doctors))
		return apperror.ErrShiftNotFound
	}

	appointmentTime := payload.AppointmentTime
	if t, err := time.Parse(time.RFC3339, payload.AppointmentTime); err == nil {
		appointmentTime = t.Format("15:04")
	}

	overlap, err := s.hasOverlappingAppointment(ctx, shift.ID, appointmentTime)
	if err != nil {
		return err
	}
	if overlap {
		return apperror.ErrShiftAlreadyBooked
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		bookingKey := cache.BookingShiftKey(shift.ID)
		lockValue, err := s.lock.Acquire(ctx, bookingKey, 5*time.Second)
		if err != nil {
			return err
		}
		if lockValue == "" {
			return apperror.ErrShiftAlreadyBooked
		}
		defer func() {
			if err := s.lock.Release(ctx, bookingKey, lockValue); err != nil {
				slog.Error("Failed to release lock booking: ", "err", err)
			}
		}()

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

func parseClockTime(s string) (time.Time, error) {
	if t, err := time.Parse("15:04:05", s); err == nil {
		return t, nil
	}
	return time.Parse("15:04", s)
}

func (s *AppointmentService) hasOverlappingAppointment(ctx context.Context, shiftID uint, appointmentTime string) (bool, error) {
	newStart, err := parseClockTime(appointmentTime)
	if err != nil {
		slog.Error("parseClockTime", "err", err)
		return false, apperror.ErrInvalidAppointmentTime
	}
	newEnd := newStart.Add(appointmentDuration)

	var existing []model.Appointment
	if err := s.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Where("shift_id = ?", shiftID).
		Where("status IN ?", []model.AppointmentStatus{
			model.AppointmentPending,
			model.AppointmentConfirmed,
			model.AppointmentCompleted,
		}).
		Find(&existing).Error; err != nil {
		return false, err
	}

	for _, a := range existing {
		existingStart, err := parseClockTime(a.AppointmentTime)
		if err != nil {
			continue
		}
		existingEnd := existingStart.Add(appointmentDuration)
		if newStart.Before(existingEnd) && existingStart.Before(newEnd) {
			return true, nil
		}
	}
	return false, nil
}

func (s *AppointmentService) ListBookedTimes(ctx context.Context, shiftID uint) ([]string, error) {
	var times []string
	err := s.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Where("shift_id = ?", shiftID).
		Where("status IN ?", []model.AppointmentStatus{
			model.AppointmentPending,
			model.AppointmentConfirmed,
			model.AppointmentCompleted,
		}).
		Pluck("appointment_time", &times).Error
	return times, err
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
