package model

import "time"

type PeriodStatus string
type ShiftStatus string

const (
	PeriodDraft     PeriodStatus = "draft"
	PeriodPublished PeriodStatus = "published"

	ShiftScheduled ShiftStatus = "scheduled"
	ShiftCancelled ShiftStatus = "cancelled"
	ShiftCompleted ShiftStatus = "completed"
)

type SchedulePeriod struct {
	ID         uint         `gorm:"primaryKey"`
	HospitalID uint         `gorm:"not null"`
	Name       string       `gorm:"not null"`
	StartDate  time.Time    `gorm:"type:date;not null"`
	EndDate    time.Time    `gorm:"type:date;not null"`
	Status     PeriodStatus `gorm:"not null;default:'draft'"`
	CreatedBy  *uint

	Shifts    []ScheduleShift `gorm:"foreignKey:PeriodID"`
	Hospital  Hospital        `gorm:"foreignKey:HospitalID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ScheduleShift struct {
	ID           uint      `gorm:"primaryKey"`
	HospitalID   uint      `gorm:"not null"`
	PeriodID     uint      `gorm:"not null"`
	DepartmentID uint      `gorm:"not null"`
	Date         time.Time `gorm:"type:date;not null"`
	TimeStart    string    `gorm:"type:varchar(5);not null"`
	TimeEnd      string    `gorm:"type:varchar(5);not null"`
	RoomNumber   string
	Status       ShiftStatus `gorm:"not null;default:'scheduled'"`
	Note         string

	Period     SchedulePeriod        `gorm:"foreignKey:PeriodID"`
	Department Department            `gorm:"foreignKey:DepartmentID"`
	Doctors    []ScheduleShiftDoctor `gorm:"foreignKey:ShiftID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ScheduleShiftDoctor struct {
	ShiftID uint `gorm:"primaryKey"`
	UserID  uint `gorm:"primaryKey"`

	User       User       `gorm:"foreignKey:UserID"`
	UserDetail UserDetail `gorm:"foreignKey:UserId;references:UserID"`
}
