package model

import "time"

type UserDetail struct {
	ID           uint
	Fullname     string `gorm:"column:full_name"`
	Address      string
	PhoneNumber  string
	DepartmentID *uint

	User       User
	UserId     uint
	Department *Department `gorm:"foreignKey:DepartmentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
