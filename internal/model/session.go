package model

import "time"

type Session struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	TokenHash string    `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`

	IP        string
	UserAgent string

	Revoked    bool `gorm:"not null;default:false"`
	RevokedAt  *time.Time
	LastUsedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}
