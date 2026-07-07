package model

import "time"

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationRevoked  InvitationStatus = "revoked"
	InvitationExpired  InvitationStatus = "expired"
)

type HospitalInvitation struct {
	ID         uint             `gorm:"primaryKey"`
	HospitalID uint             `gorm:"not null"`
	Email      string           `gorm:"not null"`
	Role       UserRole         `gorm:"not null;default:'doctor'"`
	TokenHash  string           `gorm:"not null;unique" json:"-"`
	Status     InvitationStatus `gorm:"not null;default:'pending'"`
	InvitedBy  uint             `gorm:"not null"`
	ExpiresAt  time.Time        `gorm:"not null"`

	Hospital Hospital `gorm:"foreignKey:HospitalID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
