package model

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID       uint
	Email    string
	Password string
	Role     UserRole

	CreatedAt time.Time
	UpdatedAt time.Time
}
