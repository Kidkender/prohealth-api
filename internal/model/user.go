package model

import "time"

type SystemRole string
type UserStatus string

const (
	SystemRoleAdmin SystemRole = "super_admin"
	SystemRoleUser  SystemRole = "user"

	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

type User struct {
	ID         uint       `gorm:"primaryKey"`
	Email      string     `gorm:"unique"`
	Password   string     `gorm:"not null"`
	SystemRole SystemRole `gorm:"not null;default:user"`
	UserStatus UserStatus `gorm:"not null;default:active"`

	Details *UserDetail `gorm:"foreignKey:UserId"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
