package model

import (
	"github.com/kidkender/prohealth-api/internal/llm"
)

type ChatSession struct {
	SessionID           string
	HospitalID          uint
	UserID              *uint
	PatientName         string
	PatientPhone        string
	PatientEmail        string
	DepartmentID        *uint
	Date                string
	StartTime           string
	ShiftID             *uint
	Reason              *string
	PendingConfirmation bool
	History             []llm.Message
}
