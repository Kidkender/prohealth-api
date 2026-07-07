package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/audit"
	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/lock"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/timeutil"
	"gorm.io/gorm"
)

type ScheduleService struct {
	db   *gorm.DB
	lock *lock.DistributedLock
}

func NewScheduleService(db *gorm.DB, lock *lock.DistributedLock) *ScheduleService {
	return &ScheduleService{db: db, lock: lock}
}

func (s *ScheduleService) CreateSchedulePeriod(ctx context.Context, actorID uint, hospitalID uint, payload dto.CreateScheduleInput) error {

	if payload.StartDate.After(payload.EndDate) {
		return apperror.ErrInvalidTimeRange
	}

	period := model.SchedulePeriod{
		Name:       payload.Name,
		StartDate:  payload.StartDate,
		EndDate:    payload.EndDate,
		Status:     payload.Status,
		HospitalID: hospitalID,
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&period).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionCreate, "schedule_period", period.ID)
	})
}

func (s *ScheduleService) CreateScheduleShift(ctx context.Context, actorID uint, hospitalID uint, payload dto.CreateScheduleShiftInput) error {
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
		TimeStart:    payload.StartTime.Format("15:04"),
		TimeEnd:      payload.EndTime.Format("15:04"),
		RoomNumber:   payload.RoomNumber,
		Status:       payload.Status,
		HospitalID:   hospitalID,
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&shift).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionCreate, "schedule_shift", shift.ID)
	})
}

func (s *ScheduleService) ReplaceDoctorsForShift(ctx context.Context, actorID uint, shiftID uint, doctorIDs []uint) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, shiftID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrScheduleShiftNotFound
		}
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("shift_id = ?", shiftID).Delete(&model.ScheduleShiftDoctor{}).Error; err != nil {
			return err
		}
		for _, uid := range doctorIDs {
			if err := tx.WithContext(ctx).Create(&model.ScheduleShiftDoctor{ShiftID: shiftID, UserID: uid}).Error; err != nil {
				return err
			}
		}
		return audit.Record(tx, shift.HospitalID, &actorID, model.ActionUpdate, "schedule_shift_doctors", shiftID)
	})
}

func (s *ScheduleService) SetDoctorForShift(ctx context.Context, actorID uint, payload dto.SetDoctorForShiftInput) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, payload.ShiftID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrScheduleShiftNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&model.ScheduleShiftDoctor{
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
		return audit.Record(tx, shift.HospitalID, &actorID, model.ActionCreate, "schedule_shift_doctor", payload.ShiftID)
	})
}

func (s *ScheduleService) UpdateSchedulePeriod(ctx context.Context, actorID uint, id uint, payload dto.UpdateSchedulePeriodInput) error {
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
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&period).Error; err != nil {
			return err
		}
		return audit.Record(tx, period.HospitalID, &actorID, model.ActionUpdate, "schedule_period", id)
	})
}

func (s *ScheduleService) GetSchedulePeriods(ctx context.Context, hospitalID uint, filter dto.FilterSchedulePeriod) ([]model.SchedulePeriod, int64, error) {
	var periods []model.SchedulePeriod
	var total int64

	query := s.db.WithContext(ctx).
		Model(&model.SchedulePeriod{})

	if hospitalID > 0 {
		query = query.Where("hospital_id = ?", hospitalID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartDate != nil {
		query = query.Where("start_date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("end_date <= ?", filter.EndDate)
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

func (s *ScheduleService) GetScheduleShifts(ctx context.Context, hospitalID uint, filter dto.FilterScheduleShift) ([]model.ScheduleShift, int64, error) {
	var shifts []model.ScheduleShift
	var total int64

	query := s.db.WithContext(ctx).
		Model(&model.ScheduleShift{}).
		Where("hospital_id = ?", hospitalID)

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

func (s *ScheduleService) DeleteSchedulePeriod(ctx context.Context, actorID uint, periodID uint) error {
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

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&period).Error; err != nil {
			return err
		}
		return audit.Record(tx, period.HospitalID, &actorID, model.ActionDelete, "schedule_period", periodID)
	})
}

func (s *ScheduleService) DeleteScheduleShift(ctx context.Context, actorID uint, shiftID uint) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, shiftID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrScheduleShiftNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&shift).Error; err != nil {
			return err
		}
		return audit.Record(tx, shift.HospitalID, &actorID, model.ActionDelete, "schedule_shift", shiftID)
	})
}

func (s *ScheduleService) UpdateScheduleShift(ctx context.Context, actorID uint, shiftID uint, payload dto.UpdateScheduleShiftInput) error {
	var shift model.ScheduleShift
	if err := s.db.WithContext(ctx).First(&shift, shiftID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrScheduleShiftNotFound
		}
		return err
	}

	updates := map[string]any{}

	if payload.Status != nil {
		updates["status"] = *payload.Status
	}
	if payload.RoomNumber != nil {
		updates["room_number"] = *payload.RoomNumber
	}
	if payload.Note != nil {
		updates["note"] = *payload.Note
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&model.ScheduleShift{}).
			Where("id = ?", shiftID).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return apperror.ErrScheduleShiftNotFound
		}
		return audit.Record(tx, shift.HospitalID, &actorID, model.ActionUpdate, "schedule_shift", shiftID)
	})
}

func (s *ScheduleService) GetScheduleShiftByDoctor(ctx context.Context, hospitalID uint, doctorID uint, filter *dto.FilterScheduleShift) ([]model.ScheduleShift, error) {
	var shifts []model.ScheduleShift

	query := s.db.WithContext(ctx).
		Model(&model.ScheduleShift{}).
		Where("schedule_shifts.hospital_id = ?", hospitalID).
		Joins("JOIN schedule_shift_doctors ON schedule_shift_doctors.shift_id = schedule_shifts.id").
		Where("schedule_shift_doctors.user_id = ?", doctorID).
		Preload("Department").
		Preload("Doctors.UserDetail")

	if filter.Date != nil {
		query = query.Where("schedule_shifts.date = ?", *filter.Date)
	}

	if err := query.Find(&shifts).Error; err != nil {
		return nil, err
	}
	return shifts, nil
}

func (s *ScheduleService) GenerateSchedule(
	ctx context.Context, actorID uint, hospitalID uint, payload dto.GenerateScheduleInput,
) (*dto.GenerateScheduleResponse, error) {
	key := cache.LockSechduleKey(hospitalID)

	lockValue, err := s.lock.Acquire(ctx, key, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("acquire schedule lock: %w", err)
	}
	if lockValue == "" {
		return nil, apperror.ErrSchedulePeriodGenerating
	}
	defer func() {
		if err := s.lock.Release(ctx, key, lockValue); err != nil {
			slog.Error("failed to release lock", "error", err)
		}
	}()

	monday := timeutil.NextMonday(payload.StartDate)
	friday := monday.AddDate(0, 0, 4)
	var count int64

	timeStart, err := time.Parse("15:04", payload.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	timeEnd, err := time.Parse("15:04", payload.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	if !timeEnd.After(timeStart) {
		return nil, fmt.Errorf("start time must be before end time")
	}

	var period model.SchedulePeriod

	shiftsCreated := 0
	if err := s.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.WithContext(ctx).
			Model(&model.SchedulePeriod{}).
			Where("hospital_id = ? AND start_date = ? AND status = ?", hospitalID, monday, model.PeriodPublished).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return apperror.ErrSchedulePeriodAlreadyExists
		}

		period = model.SchedulePeriod{
			HospitalID: hospitalID,
			Name:       fmt.Sprintf("Week %s - %s", monday.Format("Jan 2"), friday.Format("Jan 2, 2006")),
			StartDate:  monday,
			EndDate:    friday,
			Status:     model.PeriodPublished,
		}
		if err := tx.WithContext(ctx).Create(&period).Error; err != nil {
			return err
		}

		const maxShiftsPerWeek = 3
		const maxDeptsPerSlot = 2

		slotDeptCount := map[string]int{}
		doctorDayAssigned := map[string]map[uint]bool{}

		var depts []model.Department
		if err := tx.WithContext(ctx).
			Joins("JOIN user_details ON user_details.department_id = departments.id").
			Joins("JOIN hospital_members ON hospital_members.user_id = user_details.user_id").
			Where("hospital_members.hospital_id = ? AND hospital_members.role = ?", hospitalID, model.RoleDoctor).
			Group("departments.id").
			Find(&depts).Error; err != nil {
			return err
		}

		for deptIdx, dept := range depts {
			var userDetails []model.UserDetail
			if err := tx.WithContext(ctx).
				Joins("JOIN hospital_members ON hospital_members.user_id = user_details.user_id").
				Where("user_details.department_id = ? AND hospital_members.hospital_id = ? AND hospital_members.role = ?",
					dept.ID, hospitalID, model.RoleDoctor).
				Find(&userDetails).Error; err != nil {
				return err
			}

			roomNumber := fmt.Sprintf("Room %d%02d", deptIdx/3+1, deptIdx%3+1)
			subShifts := timeutil.SplitTimeRangeBalanced(timeStart, timeEnd)

			type slot struct {
				date time.Time
				sub  [2]time.Time
			}
			numWorkDays := 2 + rand.Intn(2)
			dayIndices := rand.Perm(5)[:numWorkDays]

			allSlots := make([]slot, 0, numWorkDays*len(subShifts))
			for _, dayIdx := range dayIndices {
				for _, sub := range subShifts {
					allSlots = append(allSlots, slot{monday.AddDate(0, 0, dayIdx), sub})
				}
			}
			rand.Shuffle(len(allSlots), func(i, j int) {
				allSlots[i], allSlots[j] = allSlots[j], allSlots[i]
			})

			doctorWeekCount := make(map[uint]int, len(userDetails))
			doctorIdx := 0

			for _, s := range allSlots {
				slotKey := s.date.Format("2006-01-02") + "|" + s.sub[0].Format("15:04") + "|" + s.sub[1].Format("15:04")
				if slotDeptCount[slotKey] >= maxDeptsPerSlot {
					continue
				}

				dateKey := s.date.Format("2006-01-02")
				if doctorDayAssigned[dateKey] == nil {
					doctorDayAssigned[dateKey] = map[uint]bool{}
				}

				assigned := false
				for attempt := 0; attempt < len(userDetails); attempt++ {
					ud := userDetails[doctorIdx%len(userDetails)]
					doctorIdx++
					if doctorWeekCount[ud.UserId] < maxShiftsPerWeek && !doctorDayAssigned[dateKey][ud.UserId] {
						shift := model.ScheduleShift{
							HospitalID:   hospitalID,
							PeriodID:     period.ID,
							DepartmentID: dept.ID,
							Date:         s.date,
							TimeStart:    s.sub[0].Format("15:04"),
							TimeEnd:      s.sub[1].Format("15:04"),
							RoomNumber:   roomNumber,
							Status:       model.ShiftScheduled,
						}
						if err := tx.WithContext(ctx).Create(&shift).Error; err != nil {
							return err
						}
						if err := tx.WithContext(ctx).Create(&model.ScheduleShiftDoctor{
							ShiftID: shift.ID,
							UserID:  ud.UserId,
						}).Error; err != nil {
							return err
						}
						doctorWeekCount[ud.UserId]++
						doctorDayAssigned[dateKey][ud.UserId] = true
						slotDeptCount[slotKey]++
						shiftsCreated++
						assigned = true
						break
					}
				}
				if !assigned {
					continue
				}
			}
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionCreate, "schedule_period_generated", period.ID)
	}); err != nil {
		return nil, err
	}

	return &dto.GenerateScheduleResponse{
		PeriodID:      period.ID,
		PeriodName:    period.Name,
		ShiftsCreated: shiftsCreated,
	}, nil

}
