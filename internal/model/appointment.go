package model

import "time"

type AppointmentStatus string
type AppointmentReason string

const (
	AppointmentPending   AppointmentStatus = "pending"
	AppointmentConfirmed AppointmentStatus = "confirmed"
	AppointmentCompleted AppointmentStatus = "completed"
	AppointmentCancelled AppointmentStatus = "cancelled"

	ReasonRoutineCheckup  AppointmentReason = "routineCheckup"
	ReasonNewPatientVisit AppointmentReason = "newPatientVisit"
	ReasonSpecificConcern AppointmentReason = "specificConcern"
)

type Appointment struct {
	ID              uint              `gorm:"primaryKey"`
	ShiftID         uint              `gorm:"not null"`
	PatientName     string            `gorm:"not null"`
	PatientPhone    string
	AppointmentTime string            `gorm:"type:varchar(8);not null"`
	Reason          AppointmentReason `gorm:"not null;default:'newPatientVisit'"`
	Status          AppointmentStatus `gorm:"not null;default:'pending'"`
	Notes           string

	Shift ScheduleShift `gorm:"foreignKey:ShiftID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
