package model

import "time"

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleDoctor UserRole = "doctor"
	RoleUser   UserRole = "user"
)

type User struct {
	ID       uint
	Email    string
	Password string
	Role     UserRole

	CreatedAt time.Time
	UpdatedAt time.Time
}
