package model

import "time"

type UserDetail struct {
	ID       uint
	Fullname string
	Avatar   string
	Address  string

	User   User
	UserId uint

	CreatedAt time.Time
	UpdatedAt time.Time
}
