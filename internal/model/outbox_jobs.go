package model

import "time"

type OutboxJobStatus string
type OutboxJobType string

const (
	OutboxJobStatusPending   OutboxJobStatus = "pending"
	OutboxJobStatusCompleted OutboxJobStatus = "completed"
	OutboxJobStatusFailed    OutboxJobStatus = "failed"
	OutboxJobStatusPublished OutboxJobStatus = "published"
)

const (
	OutboxJobTypeAppointmentCreated OutboxJobType = "email.appointment_created"
)

type OutboxJob struct {
	ID         uint            `json:"id"`
	Type       string          `json:"type"`
	Status     OutboxJobStatus `json:"status"`
	Payload    []byte          `json:"payload"`
	RetryCount int             `json:"retry_count"`
	LastError  string          `json:"last_error"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
