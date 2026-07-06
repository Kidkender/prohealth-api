package model

import "time"

type AuditActionType string

const (
	ActionCreate AuditActionType = "create"
	ActionUpdate AuditActionType = "update"
	ActionDelete AuditActionType = "delete"
)

type AuditMetadata struct {
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Before    any    `json:"before,omitempty"`
	After     any    `json:"after,omitempty"`
}

type AuditLog struct {
	ID         uint            `gorm:"primaryKey"`
	HospitalID uint            `gorm:"index"`
	ActorID    *uint           `gorm:"index"`
	Action     AuditActionType `gorm:"not null"`
	EntityType string          `gorm:"not null"`
	EntityID   uint            `gorm:"index"`
	Metadata   AuditMetadata   `gorm:"type:jsonb;serializer:json"`
	CreatedAt  time.Time
}
