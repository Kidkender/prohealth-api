package dto

import (
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
)

type CreateScheduleInput struct {
	Name      string             `json:"name" binding:"required"`
	StartDate time.Time          `json:"start_date" binding:"required"`
	EndDate   time.Time          `json:"end_date" binding:"required"`
	Status    model.PeriodStatus `json:"status" binding:"required,oneof=draft published"`
}

type CreateScheduleShiftInput struct {
	PeriodID     uint              `json:"period_id" binding:"required"`
	DepartmentID uint              `json:"department_id" binding:"required"`
	Date         time.Time         `json:"date" binding:"required"`
	StartTime    time.Time         `json:"start_time" binding:"required"`
	EndTime      time.Time         `json:"end_time" binding:"required"`
	RoomNumber   string            `json:"room_number" binding:"required"`
	Status       model.ShiftStatus `json:"status" binding:"required"`
}

type UpdateSchedulePeriodInput struct {
	Status    model.PeriodStatus `json:"status" binding:"required"`
	StartDate time.Time          `json:"start_date" binding:"required"`
	EndDate   time.Time          `json:"end_date" binding:"required"`
}

type SetDoctorForShiftInput struct {
	ShiftID  uint `json:"shift_id" binding:"required"`
	DoctorID uint `json:"doctor_id" binding:"required"`
}

type FilterSchedulePeriod struct {
	PaginationQuery
	StartDate *time.Time         `form:"start_date" `
	EndDate   *time.Time         `form:"end_date" `
	Status    model.PeriodStatus `form:"status" `
}

type FilterScheduleShift struct {
	PaginationQuery
	PeriodID     *uint              `form:"period_id" `
	DepartmentID *uint              `form:"department_id" `
	Date         *time.Time         `form:"date" `
	StartTime    *time.Time         `form:"start_time" `
	EndTime      *time.Time         `form:"end_time" `
	RoomNumber   *string            `form:"room_number" `
	Status       *model.ShiftStatus `form:"status" `
}
