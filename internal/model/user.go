package model

import "time"

type UserRole string

type User struct {
	ID       uint     `gorm:"primaryKey"`
	Email    string   `gorm:"unique"`
	Password string   `gorm:"not null"`
	Hospital Hospital `gorm:"foreignKey:HospitalID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
