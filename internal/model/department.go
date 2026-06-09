package model

import "time"

type Department struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null;unique"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
