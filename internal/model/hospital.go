package model

import "time"

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleDoctor UserRole = "doctor"
	RoleUser   UserRole = "user"
)

type Hospital struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Slug        string `gorm:"unique;not null"`
	Description string
	Address     string

	Members      []HospitalMember `gorm:"foreignKey:HospitalID"`
	Periods      []SchedulePeriod `gorm:"foreignKey:HospitalID"`
	Appointments []Appointment    `gorm:"foreignKey:HospitalID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type HospitalMember struct {
	HospitalID uint     `gorm:"primaryKey"`
	UserID     uint     `gorm:"primaryKey"`
	Role       UserRole `gorm:"not null;default:'doctor'"`
	JoinedAt   time.Time

	Hospital Hospital `gorm:"foreignKey:HospitalID"`
	User     User     `gorm:"foreignKey:UserID"`
}
