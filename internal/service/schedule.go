package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type ScheduleService struct {
	db *gorm.DB
}

func NewScheduleService(db *gorm.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

func (s *ScheduleService) CreateSchedulePeriod(ctx context.Context, payload dto.CreateScheduleInput) error {

	if payload.StartDate.After(payload.EndDate) {
		return apperror.ErrInvalidTimeRange
	}

	return s.db.WithContext(ctx).Create(&model.SchedulePeriod{
		Name:      payload.Name,
		StartDate: payload.StartDate,
		EndDate:   payload.EndDate,
		Status:    payload.Status,
	}).Error
}

func (s *ScheduleService) CreateScheduleShift(ctx context.Context, payload dto.CreateScheduleShiftInput) error {
	var period model.SchedulePeriod
	if err := s.db.WithContext(ctx).First(&period, payload.PeriodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrSchedulePeriodNotFound
		}
		return err
	}

	var department model.Department
	if err := s.db.WithContext(ctx).First(&department, payload.DepartmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrDepartmentNotFound
		}
		return err
	}

	if !payload.EndTime.After(payload.StartTime) {
		return apperror.ErrInvalidTimeRange
	}

	shift := model.ScheduleShift{
		PeriodID:     payload.PeriodID,
		DepartmentID: payload.DepartmentID,
		Date:         payload.Date,
		TimeStart:    payload.StartTime,
		TimeEnd:      payload.EndTime,
		RoomNumber:   payload.RoomNumber,
		Status:       payload.Status,
	}

	return s.db.WithContext(ctx).Create(&shift).Error
}

func (s *ScheduleService) SetDoctorForShift(ctx context.Context, payload dto.SetDoctorForShiftInput) error {
	if err := s.db.WithContext(ctx).First(&model.ScheduleShift{}, payload.ShiftID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrScheduleShiftNotFound
		}
		return err
	}

	err := s.db.WithContext(ctx).Create(&model.ScheduleShiftDoctor{
		ShiftID: payload.ShiftID,
		UserID:  payload.DoctorID,
	}).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperror.ErrDoctorAlreadyAssigned
		}
		return err
	}
	return nil
}

func (s *ScheduleService) UpdateSchedulePeriod(ctx context.Context, id uint, payload dto.UpdateSchedulePeriodInput) error {
	var period model.SchedulePeriod
	if err := s.db.WithContext(ctx).First(&period, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrSchedulePeriodNotFound
		}
		return err
	}

	if !payload.EndDate.After(payload.StartDate) {
		return apperror.ErrInvalidTimeRange
	}

	period.Status = payload.Status
	period.StartDate = payload.StartDate
	period.EndDate = payload.EndDate
	return s.db.Save(&period).Error
}

func (s *ScheduleService) GetSchedulePeriods(ctx context.Context, filter dto.FilterSchedulePeriod) ([]model.SchedulePeriod, int64, error) {
	var periods []model.SchedulePeriod
	var total int64
	query := s.db.Model(&model.SchedulePeriod{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.StartDate != nil {
		query = query.Where("start_date>=?", filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("end_date<=?", filter.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(filter.Limit).
		Offset(filter.Offset()).
		Find(&periods).Error; err != nil {
		return nil, 0, err
	}

	return periods, total, nil
}

func (s *ScheduleService) GetSchedulePeriodById(ctx context.Context, id uint) (*model.SchedulePeriod, error) {
	var period model.SchedulePeriod
	if err := s.db.WithContext(ctx).First(&period, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrSchedulePeriodNotFound
		}
		return nil, err
	}
	return &period, nil
}

func (s *ScheduleService) GetScheduleShifts(ctx context.Context, filter dto.FilterScheduleShift) ([]model.ScheduleShift, int64, error) {
	var shifts []model.ScheduleShift
	var total int64

	query := s.db.Model(&model.ScheduleShift{})
	if filter.PeriodID != nil {
		query = query.Where("period_id = ?", *filter.PeriodID)
	}
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", *filter.DepartmentID)
	}
	if filter.Date != nil {
		query = query.Where("date = ?", *filter.Date)
	}
	if filter.StartTime != nil {
		query = query.Where("time_start = ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("time_end = ?", *filter.EndTime)
	}
	if filter.RoomNumber != nil {
		query = query.Where("room_number = ?", *filter.RoomNumber)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(filter.Limit).
		Offset(filter.Offset()).
		Find(&shifts).Error; err != nil {
		return nil, 0, err
	}
	return shifts, total, nil
}

func (s *ScheduleService) GetScheduleShiftsByPeriodID(ctx context.Context, periodID uint) ([]model.ScheduleShift, error) {
	var shifts []model.ScheduleShift
	if err := s.db.WithContext(ctx).
		Where("period_id = ?", periodID).
		Preload("Department").
		Preload("Doctors.UserDetail").
		Find(&shifts).Error; err != nil {
		return nil, err
	}
	return shifts, nil
}

func (s *ScheduleService) GetScheduleShiftById(ctx context.Context, id uint) (*model.ScheduleShift, error) {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrScheduleShiftNotFound
		}
		return nil, err
	}
	return &shift, nil
}

func (s *ScheduleService) DeleteSchedulePeriod(ctx context.Context, periodID uint) error {
	var period model.SchedulePeriod

	if err := s.db.WithContext(ctx).First(&period, periodID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrSchedulePeriodNotFound
		}
		return err
	}

	var shiftCount int64
	s.db.Model(&model.ScheduleShift{}).Where("period_id = ?", periodID).Count(&shiftCount)
	if shiftCount > 0 {
		return apperror.ErrSchedulePeriodHasShifts
	}

	return s.db.WithContext(ctx).Delete(&period).Error
}

func (s *ScheduleService) DeleteScheduleShift(ctx context.Context, shiftID uint) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, shiftID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrScheduleShiftNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Delete(&shift).Error
}
