package model

import "time"

type SystemRole string

const (
	SystemRoleAdmin SystemRole = "super_admin"
	SystemRoleUser  SystemRole = "user"
)

type User struct {
	ID         uint       `gorm:"primaryKey"`
	Email      string     `gorm:"unique"`
	Password   string     `gorm:"not null"`
	SystemRole SystemRole `gorm:"not null;default:user"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
